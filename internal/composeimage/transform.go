// Package composeimage rewrites a Compose file for artifacts deploy (build → image).
package composeimage

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Options controls GenerateForArtifacts.
type Options struct {
	// TargetService is the app service whose build: block is replaced when AllBuilt is false (required then).
	TargetService string
	// ImageExpr is the YAML value for image: when ServiceImages is nil (default ${DEPLOY_IMAGE}).
	ImageExpr string
	// ServiceImages maps service name → full image reference for AllBuilt mode. When non-nil,
	// every service with build: must appear in the map; each gets that literal image: value.
	// When nil and AllBuilt, ImageExpr is used for all built services (same tag).
	ServiceImages map[string]string
	// AllBuilt if true, every service with build: is converted to image: (same ImageExpr).
	// If false, only TargetService is converted and any other service with build: is an error.
	AllBuilt bool
}

// GenerateForArtifacts reads a Compose YAML document and produces a variant where built services
// use image: instead of build:.
func GenerateForArtifacts(composeYAML []byte, o Options) ([]byte, error) {
	target := strings.TrimSpace(o.TargetService)
	if !o.AllBuilt && target == "" {
		return nil, fmt.Errorf("target service name is empty")
	}
	img := strings.TrimSpace(o.ImageExpr)
	if img == "" {
		img = "${DEPLOY_IMAGE}"
	}

	var root map[string]interface{}
	if err := yaml.Unmarshal(composeYAML, &root); err != nil {
		return nil, fmt.Errorf("parse compose yaml: %w", err)
	}
	if root == nil {
		return nil, fmt.Errorf("compose file is empty")
	}

	svcObj, ok := root["services"]
	if !ok || svcObj == nil {
		return nil, fmt.Errorf("no services: key in compose file")
	}
	services, ok := svcObj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("services must be a mapping")
	}

	if !o.AllBuilt {
		var otherBuilt []string
		for name, raw := range services {
			if raw == nil {
				continue
			}
			svc, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			if !hasBuild(svc) {
				continue
			}
			if name == target {
				continue
			}
			otherBuilt = append(otherBuilt, name)
		}
		if len(otherBuilt) > 0 {
			return nil, fmt.Errorf("services still declare build (only %q should): %s", target, strings.Join(otherBuilt, ", "))
		}

		tgtRaw, ok := services[target]
		if !ok || tgtRaw == nil {
			return nil, fmt.Errorf("service %q not found under services:", target)
		}
		targetSvc, ok := tgtRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("service %q must be a mapping", target)
		}
		if !hasBuild(targetSvc) {
			return nil, fmt.Errorf("service %q has no build: — nothing to convert for artifacts", target)
		}
	}

	out := cloneMapStringInterface(root)
	outSvcs := out["services"].(map[string]interface{})

	if o.AllBuilt {
		perSvc := o.ServiceImages
		converted := 0
		for name, raw := range outSvcs {
			if raw == nil {
				continue
			}
			svc, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			if !hasBuild(svc) {
				continue
			}
			var imageVal string
			if len(perSvc) > 0 {
				ref, ok := perSvc[name]
				if !ok {
					return nil, fmt.Errorf("deploy_images missing entry for service %q (every built service needs an image tag)", name)
				}
				ref = strings.TrimSpace(ref)
				if ref == "" {
					return nil, fmt.Errorf("deploy_images[%q] is empty", name)
				}
				imageVal = ref
			} else {
				imageVal = img
			}
			cp := cloneMapStringInterface(svc)
			delete(cp, "build")
			cp["image"] = imageVal
			outSvcs[name] = cp
			converted++
		}
		if converted == 0 {
			return nil, fmt.Errorf("no service with build: found")
		}
	} else {
		outTarget := cloneMapStringInterface(outSvcs[target].(map[string]interface{}))
		delete(outTarget, "build")
		outTarget["image"] = img
		outSvcs[target] = outTarget
	}

	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(out); err != nil {
		return nil, err
	}
	_ = enc.Close()
	return []byte(buf.String()), nil
}

func hasBuild(svc map[string]interface{}) bool {
	v, ok := svc["build"]
	if !ok {
		return false
	}
	return v != nil
}

func cloneMapStringInterface(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = deepCloneValue(v)
	}
	return out
}

func deepCloneValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		return cloneMapStringInterface(x)
	case []interface{}:
		cp := make([]interface{}, len(x))
		for i, e := range x {
			cp[i] = deepCloneValue(e)
		}
		return cp
	default:
		return v
	}
}
