package interpret

type Environment struct {
	Address     Address         `json:"address"`
	Os          string          `json:"os"`
	Provider    string          `json:"provider"`
	Type        string          `json:"type"`
	Nick        string          `json:"nick"`
	LogicStates []LogicState    `json:"logic-states"`
	Datasets    []ParentDataset `json:"datasets"`
	Processes   []Process       `json:"processes"`
}

type EnvLink struct {
	Address Address `json:"address"`
	Nick    string  `json:"nick"`
}

type Address struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
}

type LogicState struct {
	Datasets    []DataLink `json:"datasets"`
	Environment EnvLink    `json:"environment"`
	ID          struct {
		Commit    []byte
		CommitStr string `json:"commit"`
		Version   string `json:"version"`
		Semver    string `json:"semver"`
	} `json:"id"`
	Lgroup string `json:"lgroup"`
	Nick   string `json:"nick"`
	Path   string `json:"path"`
	Type   string `json:"type"`
}

type DataLink struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Subset      string   `json:"subset"`
	Interaction string   `json:"interaction"`
	ConnUnix    ConnUnix `json:"connUnix"`
	ConnNet     ConnNet  `json:"connNet"`
}

type ConnNet struct {
	Hostname string `json:"hostname"`
	Ipv4     string `json:"ipv4"`
	Ipv6     string `json:"ipv6"`
	Port     int    `json:"port"`
	Proto    string `json:"proto"`
}

type ConnUnix struct {
	Path string `json:"path"`
}

type CommitMeta struct {
	Sha1      []byte
	Sha1Str   string   `json:"sha1"`
	Tags      []string `json:"tags"`
	TestState string   `json:"testState"`
}

type Commit struct {
	Author     string `json:"author"`
	Date       string `json:"date"`
	Parents    [][]byte
	ParentsStr []string `json:"parents"`
	Sha1       []byte
	Sha1Str    string `json:"sha1"`
	Subject    string `json:"subject"`
	Repository string `json:"repository"`
}

// FIXME this metaset/set design is not recursive, but it will need to be
type ParentDataset struct {
	Environment EnvLink   `json:"environment"`
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Subsets     []Dataset `json:"subsets"`
}

type Dataset struct {
	Name       string      `json:"name"`
	Parent     string      // TODO yechhhh...how do we qualify the hierarchy?
	CreateTime string      `json:"create-time"`
	Genesis    DataGenesis `json:"genesis"`
}

type DataGenesis interface {
	_dg() // dummy method, avoid propagating the interface
}

// TODO this form implies a snap that only existed while 'in flight' - that is, no
// dump artfiact that exists on disk anywhere. Need to incorporate that case, though.
type DataProvenance struct {
	Address  Address  `json:"address"`
	Dataset  []string `json:"dataset"`
	SnapTime string   `json:"snap-time"`
}

type DataAlpha string

func (d DataAlpha) _dg()      {}
func (d DataProvenance) _dg() {}

type Process struct {
	Pid         int          `json:"pid"`
	Cwd         string       `json:"cwd"`
	Environment EnvLink      `json:"environment"`
	Group       string       `json:"group"`
	Listen      []ListenAddr `json:"listen"`
	LogicStates []string     `json:"logic-states"`
	User        string       `json:"user"`
}

type ListenAddr struct {
	Port  int      `json:"port"`
	Proto []string `json:"proto"`
	Type  string   `json:"type"`
	Path  string   `json:"path"`
}
