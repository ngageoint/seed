package objects

//Seed represents a seed.manifest.json object.
type Seed struct {
	SeedVersion string `json:"seedVersion"`
	Job         struct {
		Name             string `json:"name"`
		AlgorithmVersion string `json:"algorithmVersion"`
		PackageVersion   string `json:"packageVersion"`
		Interface        struct {
			Cmd       string `json:"cmd"`
			InputData struct {
				Files []struct {
					Name      string   `json:"name"`
					MediaType []string `json:"mediaType"`
					Pattern   string   `json:"pattern"`
				}
			}
			OutputData struct {
				Files []struct {
					Name      string `json:"name"`
					MediaType string `json:"mediaType"`
					Pattern   string `json:"pattern"`
					Count     string `json:"count"`
					Required  bool   `json:"required,string"`
				}
				JSON []struct {
					Name     string `json:"name"`
					Type     string `json:"type"`
					Key      string `json:"key"`
					Required bool   `json:"required"`
				}
			}
			Mounts []struct {
				Name string `json:"name"`
				Path string `json:"path"`
				Mode string `json:"mode"`
			}
			Settings []struct {
				Name   string `json:"name"`
				Secret bool   `json:"secret"`
			}
			ErrorMapping []struct {
				Code        int    `json:"code,string"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Category    string `json:"category"`
			}
		}
	}
}
