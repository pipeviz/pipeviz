package interpret

// FIXME this metaset/set design is not recursive, but it will need to be
type ParentDataset struct {
	Environment EnvLink   `json:"environment,omitempty"`
	Path        string    `json:"path,omitempty"`
	Name        string    `json:"name,omitempty"`
	Subsets     []Dataset `json:"subsets,omitempty"`
}

type Dataset struct {
	Name       string      `json:"name,omitempty"`
	Parent     string      // TODO yechhhh...how do we qualify the hierarchy?
	CreateTime string      `json:"create-time,omitempty"`
	Genesis    DataGenesis `json:"genesis,omitempty"`
}

type DataGenesis interface {
	_dg() // dummy method, avoid propagating the interface
}

// TODO this form implies a snap that only existed while 'in flight' - that is, no
// dump artfiact that exists on disk anywhere. Need to incorporate that case, though.
type DataProvenance struct {
	Address  Address  `json:"address,omitempty"`
	Dataset  []string `json:"dataset,omitempty"`
	SnapTime string   `json:"snap-time,omitempty"`
}

type DataAlpha string

func (d DataAlpha) _dg()      {}
func (d DataProvenance) _dg() {}

type YumPkg struct {
	Name       string `json:"name,omitempty"`
	Repository string `json:"repository,omitempty"`
	Version    string `json:"version,omitempty"`
	Epoch      int    `json:"epoch,omitempty"`
	Release    string `json:"release,omitempty"`
	Arch       string `json:"arch,omitempty"`
}
