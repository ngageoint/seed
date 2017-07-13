package objects

//Seed represents a seed.manifest.json object.
type Seed struct {
	ManifestVersion string `json:"manifest_version"`
	Job             struct {
		Name             string `json:"name"`
		AlgorithmVersion string `json:"algorithmVersion"`
		PackageVersion   string `json:"packageVersion"`
		Interface        struct {
			Args      string `json:"args"`
			InputData struct {
				Files []struct {
					Name      string   `json:"name"`
					MediaType []string `json:"mediaType"`
					Pattern   string   `json:"pattern"`
					Path      string   `json:"path"`
				}
			}
			OutputData struct {
				Files []struct {
					Name      string `json:"name"`
					MediaType string `json:"mediaType"`
					Pattern   string `json:"pattern"`
					Count     string `json:"count"`
					Required  bool   `json:"required"`
				}
				Json []struct {
					Name     string `json:"name"`
					Type     string `json:"type"`
					Key      string `json:"key"`
					Required bool   `json:"required"`
				}
			}
			EnvVars []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			}
			Mounts []struct {
				Name string `json:"name"`
				Path string `json:"path"`
				Mode string `json:"mode"`
			}
			Settings []struct {
				Name   string `json:"name"`
				Secret string `json:"secret"`
			}
			ErrorMapping []struct {
				Code        int    `json:"code"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Category    string `json:"category"`
			}
		}
	}
}
