package check

type Config struct {
	CDIRfile    string `json:"cdir-file"`
	TestHost    string `json:"test-host"`
	Workers     int    `json:"workers"`
	HTTPworkers int    `json:"http-workers"`
	Ports       []int  `json:"ports"`
}
