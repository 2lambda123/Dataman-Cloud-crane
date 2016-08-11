package dockerclient

import (
	"fmt"
	"regexp"

	"github.com/Dataman-Cloud/rolex/src/util/rolexerror"

	distreference "github.com/docker/distribution/reference"
	"github.com/docker/engine-api/types/swarm"
	"github.com/docker/swarmkit/manager/scheduler"
	"github.com/docker/swarmkit/protobuf/ptypes"
)

var isValidName = regexp.MustCompile(`^[a-zA-Z0-9](?:[-_]*[A-Za-z0-9]+)*$`)

func validateResources(r *swarm.Resources) error {
	if r == nil {
		return nil
	}

	var errMsg string
	if r.NanoCPUs != 0 && r.NanoCPUs < 1e6 {
		errMsg = fmt.Sprintf("invalid cpu value %g: Must be at least %g", float64(r.NanoCPUs)/1e9, 1e6/1e9)
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceNanoCPUs, errMsg)
	}

	if r.MemoryBytes != 0 && r.MemoryBytes < 4*1024*1024 {
		errMsg = fmt.Sprintf("invalid memory value %d: Must be at least 4MiB", r.MemoryBytes)
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceMemoryBytes, errMsg)
	}
	return nil
}

func validateResourceRequirements(r *swarm.ResourceRequirements) error {
	if r == nil {
		return nil
	}
	if err := validateResources(r.Limits); err != nil {
		return err
	}
	if err := validateResources(r.Reservations); err != nil {
		return err
	}
	return nil
}

func validateRestartPolicy(rp *swarm.RestartPolicy) error {
	if rp == nil {
		return nil
	}

	var errMsg string
	if rp.Delay != nil {
		delay, err := ptypes.Duration(ptypes.DurationProto(*rp.Delay))
		if err != nil {
			return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceDelay, err.Error())
		}
		if delay < 0 {
			errMsg = "TaskSpec: restart-delay cannot be negative"
			return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceDelay, errMsg)
		}
	}

	if rp.Window != nil {
		win, err := ptypes.Duration(ptypes.DurationProto(*rp.Window))
		if err != nil {
			return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceWindow, err.Error())
		}
		if win < 0 {
			errMsg = "TaskSpec: restart-window cannot be negative"
			return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceWindow, errMsg)
		}
	}

	return nil
}

func validatePlacement(placement *swarm.Placement) error {
	if placement == nil {
		return nil
	}
	_, err := scheduler.ParseExprs(placement.Constraints)
	return rolexerror.NewRolexError(rolexerror.CodeInvalidServicePlacement, err.Error())
}

func validateUpdate(uc *swarm.UpdateConfig) error {
	if uc == nil {
		return nil
	}

	delay, err := ptypes.Duration(ptypes.DurationProto(uc.Delay))
	if err != nil {
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceDelay, err.Error())
	}

	if delay < 0 {
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceUpdateConfig, "TaskSpec: update-delay cannot be negative")
	}

	return nil
}

func validateTask(taskSpec swarm.TaskSpec) error {
	if err := validateResourceRequirements(taskSpec.Resources); err != nil {
		return err
	}

	if err := validateRestartPolicy(taskSpec.RestartPolicy); err != nil {
		return err
	}

	if err := validatePlacement(taskSpec.Placement); err != nil {
		return err
	}

	//TODO add this validate as soon
	//if taskSpec.GetRuntime() == nil {
	//	return grpc.Errorf(codes.InvalidArgument, "TaskSpec: missing runtime")
	//}

	//_, ok := taskSpec.GetRuntime().(*api.TaskSpec_Container)
	//if !ok {
	//	return grpc.Errorf(codes.Unimplemented, "RuntimeSpec: unimplemented runtime in service spec")
	//}

	//container := taskSpec.GetContainer()
	//if container == nil {
	//	return grpc.Errorf(codes.InvalidArgument, "ContainerSpec: missing in service spec")
	//}

	//if container.Image == "" {
	//	return grpc.Errorf(codes.InvalidArgument, "ContainerSpec: image reference must be provided")
	//}

	//if _, _, err := reference.Parse(container.Image); err != nil {
	//	return grpc.Errorf(codes.InvalidArgument, "ContainerSpec: %q is not a valid repository/tag", container.Image)
	//}
	return nil
}

func validateEndpointSpec(epSpec *swarm.EndpointSpec) error {
	// Endpoint spec is optional
	if epSpec == nil {
		return nil
	}

	if len(epSpec.Ports) > 0 && epSpec.Mode == swarm.ResolutionModeDNSRR {
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceEndpoint, "EndpointSpec: ports can't be used with dnsrr mode")
	}

	portSet := make(map[swarm.PortConfig]struct{})
	for _, port := range epSpec.Ports {
		if _, ok := portSet[port]; ok {
			return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceEndpoint, "EndpointSpec: duplicate ports provided")
		}

		portSet[port] = struct{}{}
	}

	return nil
}

func validateServiceSpec(spec *swarm.ServiceSpec) error {
	if spec == nil {
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceSpec, "service spec must not null")
	}
	if err := validateAnnotations(spec.Annotations); err != nil {
		return err
	}
	if err := validateTask(spec.TaskTemplate); err != nil {
		return err
	}
	if err := validateUpdate(spec.UpdateConfig); err != nil {
		return err
	}
	if err := validateEndpointSpec(spec.EndpointSpec); err != nil {
		return err
	}

	if err := validateImageName(spec.TaskTemplate.ContainerSpec.Image); err != nil {
		return err
	}
	return nil
}

func validateAnnotations(m swarm.Annotations) error {
	if m.Name == "" {
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceName, "meta: name must be provided")
	} else if !isValidName.MatchString(m.Name) {
		// if the name doesn't match the regex
		return rolexerror.NewRolexError(rolexerror.CodeInvalidServiceName, "invalid name, only [a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]")
	}
	return nil
}

func validateImageName(imageName string) error {
	_, err := distreference.ParseNamed(imageName)
	if err != nil {
		return rolexerror.NewRolexError(rolexerror.CodeInvalidImageName, err.Error())
	}
	return nil
}
