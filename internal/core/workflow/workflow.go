package workflow

type Step struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type Workflow struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Steps       []Step `yaml:"steps"`
}
