package config

import "strings"

// HasDeployImageOrImages is true if deploy_image is non-empty or deploy_images has
// at least one non-empty tag. Use before requiring an image for artifacts deploy.
func (c *Config) HasDeployImageOrImages() bool {
	if c == nil {
		return false
	}
	if strings.TrimSpace(c.DeployImage) != "" {
		return true
	}
	for _, v := range c.DeployImages {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

// ParseDeployImagesList parses "svc=image,svc2=image2" (comma between pairs) for
// the DEPLOY_IMAGES variable in dq.env or process environment (see §14.1).
// Image refs can contain colons; only the first = in each pair is the separator.
func ParseDeployImagesList(s string) map[string]string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	out := make(map[string]string)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		before, after, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		before, after = strings.TrimSpace(before), strings.TrimSpace(after)
		if before != "" && after != "" {
			out[before] = after
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
