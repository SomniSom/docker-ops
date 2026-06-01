package cli

// buildSystemPruneArgs builds argv for `docker system prune`.
func buildSystemPruneArgs(allImages, volumes bool) []string {
	args := []string{"system", "prune", "-f"}
	if allImages {
		args = append(args, "-a")
	}
	if volumes {
		args = append(args, "--volumes")
	}
	return args
}
