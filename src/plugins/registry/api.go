package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Dataman-Cloud/rolex/src/plugins/auth"
	"github.com/Dataman-Cloud/rolex/src/plugins/auth/authenticators"
	"github.com/Dataman-Cloud/rolex/src/util/config"
	"github.com/Dataman-Cloud/rolex/src/util/db"
	"github.com/Dataman-Cloud/rolex/src/util/rolexerror"
	"github.com/Dataman-Cloud/rolex/src/util/rolexgin"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/mattes/migrate/driver/mysql"
)

const manifestPattern = `^application/vnd.docker.distribution.manifest.v\d`

type Registry struct {
	Config        *config.Config
	DbClient      *gorm.DB
	Authenticator auth.Authenticator
}

func NewRegistry(Config *config.Config) *Registry {
	registry := &Registry{Config: Config, DbClient: db.DB()}

	if registry.Config.AccountAuthenticator == "db" {
		registry.Authenticator = authenticators.NewDBAuthenticator()
	} else if registry.Config.AccountAuthenticator == "ldap" {
	} else {
		registry.Authenticator = authenticators.NewDefaultAuthenticator()
	}

	return registry
}

func (registry *Registry) MigriateTable() {
	registry.DbClient.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Image{})
	registry.DbClient.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Tag{})
	registry.DbClient.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&ImageAccess{})
}

func (registry *Registry) Token(ctx *gin.Context) {
	username, password, _ := ctx.Request.BasicAuth()
	authenticated := registry.Authenticate(username, password)

	service := ctx.Query("service")
	scope := ctx.Query("scope")

	if len(scope) == 0 && !authenticated {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	accesses := registry.ParseResourceActions(scope)
	for _, access := range accesses {
		registry.FilterAccess(username, authenticated, access)
	}

	//create token
	rawToken, err := registry.MakeToken(registry.Config, username, service, accesses)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": rawToken})
}

func (registry *Registry) Authenticate(principal, password string) bool {
	_, err := registry.Authenticator.Login(&auth.Account{Email: principal, Password: password})
	if err != nil {
		return false
	}
	return true
}

func (registry *Registry) Notifications(ctx *gin.Context) {
	notification := &Notification{}
	if err := ctx.BindJSON(&notification); err != nil {
		switch jsonErr := err.(type) {
		case *json.SyntaxError:
			log.Error("Notification JSON syntax error at byte %v: %s", jsonErr.Offset, jsonErr.Error())
		case *json.UnmarshalTypeError:
			log.Error("Unexpected type at by type %v. Expected %s but received %s.",
				jsonErr.Offset, jsonErr.Type, jsonErr.Value)
		}
	}

	for _, e := range notification.Events {
		matched, _ := regexp.MatchString(manifestPattern, e.Target.MediaType)
		if matched && strings.HasPrefix(ctx.Request.UserAgent(), "Go-http-client") {
			registry.HandleNotification(e)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{})
}

func (registry *Registry) TagList(ctx *gin.Context) {
	var tags []*Tag
	err := registry.DbClient.Where("namespace = ? AND image = ?", ctx.Param("namespace"), ctx.Param("image")).Find(&tags).Error
	if err != nil {
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryTagsListError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	for _, tag := range tags {
		registry.DbClient.Model(&ImageAccess{}).Where("namespace = ? AND image = ? AND digest = ? AND action='pull'", tag.Namespace, tag.Image, tag.Digest).Count(&tag.PullCount)
		registry.DbClient.Model(&ImageAccess{}).Where("namespace = ? AND image = ? AND digest = ? AND action='push'", tag.Namespace, tag.Image, tag.Digest).Count(&tag.PushCount)
	}

	rolexgin.HttpOkResponse(ctx, tags)
}

func (registry *Registry) GetManifests(ctx *gin.Context) {
	account_, found := ctx.Get("account")
	if !found {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	account := account_.(auth.Account)

	resp, _, err := registry.RegistryAPIGet(fmt.Sprintf("%s/%s/manifests/%s", ctx.Param("namespace"), ctx.Param("image"), ctx.Param("reference")), account.Email)
	if err != nil {
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryGetManifestError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	var manifest schema1.Manifest
	err = json.Unmarshal(resp, &manifest)
	if err != nil {
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryManifestParseError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	rolexgin.HttpOkResponse(ctx, manifest)
}

func (registry *Registry) DeleteManifests(ctx *gin.Context) {
	account_, found := ctx.Get("account")
	if !found {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	account := account_.(auth.Account)

	_, _, err := registry.RegistryAPIDeleteSchemaV2(fmt.Sprintf("%s/%s/manifests/%s", ctx.Param("namespace"), ctx.Param("image"), ctx.Param("reference")), account.Email)
	if err != nil {
		err := rolexerror.NewRolexError(rolexerror.CodeRegistryManifestDeleteError, err.Error())
		rolexgin.HttpErrorResponse(ctx, err)
		return
	}

	rolexgin.HttpOkResponse(ctx, "success")
}

func (registry *Registry) MineRepositories(ctx *gin.Context) {
	keywords := ctx.Query("keywords")
	account_, found := ctx.Get("account")
	if !found {
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	account := account_.(auth.Account)

	var images []*Image
	var err error
	if len(keywords) > 0 {
		err = registry.DbClient.Where("namespace = ? AND (namespace like ? OR image like ?)", RegistryNamespaceForAccount(account),
			LikeParam(keywords), LikeParam(keywords)).Order("created_at DESC").Find(&images).Error
	} else {
		err = registry.DbClient.Where("namespace = ?", RegistryNamespaceForAccount(account)).Order("created_at DESC").Find(&images).Error
	}
	if err != nil {
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryCatalogListError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	for _, image := range images {
		registry.DbClient.Model(&ImageAccess{}).Where("namespace = ? AND image = ? AND action='pull'", image.Namespace, image.Image).Count(&image.PullCount)
		registry.DbClient.Model(&ImageAccess{}).Where("namespace = ? AND image = ? AND action='push'", image.Namespace, image.Image).Count(&image.PushCount)
	}

	rolexgin.HttpOkResponse(ctx, images)
}

func (registry *Registry) PublicRepositories(ctx *gin.Context) {
	keywords := ctx.Query("keywords")
	var images []*Image
	var err error
	if len(keywords) > 0 {
		err = registry.DbClient.Where("Publicity = 1 AND (namespace like ? OR image like ?)", LikeParam(keywords), LikeParam(keywords)).Order("created_at DESC").Find(&images).Error
	} else {
		err = registry.DbClient.Where("Publicity = 1").Order("created_at DESC").Find(&images).Error
	}
	if err != nil {
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryCatalogListError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	for _, image := range images {
		registry.DbClient.Model(&ImageAccess{}).Where("namespace = ? AND image = ? AND action='pull'", image.Namespace, image.Image).Count(&image.PullCount)
		registry.DbClient.Model(&ImageAccess{}).Where("namespace = ? AND image = ? AND action='push'", image.Namespace, image.Image).Count(&image.PushCount)
	}

	rolexgin.HttpOkResponse(ctx, images)
}

func (registry *Registry) ImagePublicity(ctx *gin.Context) {
	var param struct {
		Publicity uint8 `json:"Publicity"`
	}

	if err := ctx.BindJSON(&param); err != nil {
		switch jsonErr := err.(type) {
		case *json.SyntaxError:
			log.Error("Notification JSON syntax error at byte %v: %s", jsonErr.Offset, jsonErr.Error())
		case *json.UnmarshalTypeError:
			log.Error("Unexpected type at by type %v. Expected %s but received %s.",
				jsonErr.Offset, jsonErr.Type, jsonErr.Value)
		}
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryImagePublicityParamError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	var image Image
	registry.DbClient.Where("namespace = ? AND image = ? ", ctx.Param("namespace"), ctx.Param("image")).Find(&image)
	err := registry.DbClient.Model(&image).UpdateColumn("Publicity", param.Publicity).Error
	if err != nil {
		rolexerr := rolexerror.NewRolexError(rolexerror.CodeRegistryImagePublicityUpdateError, err.Error())
		rolexgin.HttpErrorResponse(ctx, rolexerr)
		return
	}

	rolexgin.HttpOkResponse(ctx, "success")
}

func (registry *Registry) HandleNotification(n Event) {
	if n.Action == "push" {
		if len(strings.Split(n.Target.Repository, "/")) != 2 {
			return
		}
		namespace := strings.Split(n.Target.Repository, "/")[0]
		image := strings.Split(n.Target.Repository, "/")[1]

		// create or update an image
		var modelImage Image
		err := registry.DbClient.Where("namespace = ? AND image = ?", namespace, image).Find(&modelImage).Error
		if err != nil && strings.Contains(err.Error(), "not found") {
			modelImage.Namespace = namespace
			modelImage.Image = image
			if namespace == "library" {
				modelImage.Publicity = 1
			}
		}

		resp, _, err := registry.RegistryAPIGet(fmt.Sprintf("%s/%s/tags/list", namespace, image), n.Actor.Name)
		if err != nil {
			return
		}
		var respBody struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		}
		err = json.Unmarshal(resp, &respBody)
		if err != nil {
			return
		}

		modelImage.LatestTag = respBody.Tags[0]
		if modelImage.ID != 0 {
			registry.DbClient.Model(&modelImage).Updates(Image{LatestTag: modelImage.LatestTag})
		} else {
			registry.DbClient.Save(&modelImage)
		}

		// create or update tag
		for _, t := range respBody.Tags {
			tag := &Tag{}
			err = registry.DbClient.Where("namespace = ? AND image = ? AND tag = ? ", namespace, image, t).Find(tag).Error
			if err != nil && strings.Contains(err.Error(), "not found") {
				tag.Namespace = namespace
				tag.Image = image
				tag.Tag = t
				registry.DbClient.Save(tag)
			}

			tag.Size, tag.Digest, err = registry.SizeAndReferenceForTag(fmt.Sprintf("%s/%s/manifests/%s", namespace, image, t), n.Actor.Name)
			if tag.ID != 0 {
				registry.DbClient.Model(tag).Updates(Tag{Size: tag.Size, Digest: tag.Digest})
			} else {
				registry.DbClient.Save(tag)
			}
		}
	}
	registry.LogImageAccess(n)
}

func (registry *Registry) LogImageAccess(n Event) {
	ia := &ImageAccess{}
	if len(strings.Split(n.Target.Repository, "/")) == 2 {
		ia.Namespace = strings.Split(n.Target.Repository, "/")[0]
		ia.Image = strings.Split(n.Target.Repository, "/")[1]
		ia.Digest = n.Target.Digest
		ia.AccountEmail = n.Actor.Name
		ia.Action = n.Action
		registry.DbClient.Save(ia)
	}
}

func (registry *Registry) SizeAndReferenceForTag(url string, account string) (uint64, string, error) {
	var size uint64
	content, digest, err := registry.RegistryAPIGetSchemaV2(url, account)
	if err != nil {
		return 0, "", err
	}

	var v2response V2RegistryResponse
	err = json.Unmarshal(content, &v2response)
	if err != nil {
		return 0, "", err
	}

	size = v2response.Config.Size
	for _, layer := range v2response.Layers {
		size = size + layer.Size
	}

	return size, digest, nil
}

func LikeParam(like string) string {
	return "%" + like + "%"
}
