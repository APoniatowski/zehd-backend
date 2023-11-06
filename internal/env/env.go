package env

import "os"

// EnvProfiler Retrieve the environment variable (PROFILER) and return the bool, as string, value for further processing
func EnvProfiler() string {
	profiler := os.Getenv("PROFILER")
	if len(profiler) == 0 {
		profiler = "false"
	}
	return profiler
}
