package interpret

import (
	"bytes"
	"encoding/hex"
)

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
		Commit    Sha1
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

type Sha1 [20]byte

func (s Sha1) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 42))

	_, err := buf.WriteString(`"`)
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(hex.EncodeToString(s[:]))
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(`"`)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// IsEmpty checks to see if the Sha1 is equal to the null Sha1 (20 zero bytes/40 ASCII zeroes)
func (s Sha1) IsEmpty() bool {
	// TODO this may have some odd semantic side effects, as git uses this, the "null sha1", to indicate a nonexistent head. not in any places pipeviz will forseeably interact with it, though...
	return s == [20]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
}

// String converts a sha1 to the standard 40-char lower-case hexadecimal representation.
func (s Sha1) String() string {
	return hex.EncodeToString(s[:])
}

type CommitMeta struct {
	Sha1      Sha1
	Sha1Str   string   `json:"sha1"`
	Tags      []string `json:"tags"`
	Branches  []string `json:"branches"`
	TestState string   `json:"testState"`
}

type Commit struct {
	Author     string `json:"author"`
	Date       string `json:"date"`
	Parents    []Sha1
	ParentsStr []string `json:"parents"`
	Sha1       Sha1
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
	Dataset     string       `json:"dataset"`
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
