package types2

type Named struct {
	name    string
	UT      Type
	pkgName string
}

var Error = &Named{
	name: "error",
	UT: &Interface{
		Methods: []*Func{
			&Func{
				Name: "Error",
				Typ: &Signature{
					Results: &Tuple{
						Types: []Type{String},
					},
				},
			},
		},
	},
}

func NewNamed(name string, pkgName string, typ Type) *Named {
	return &Named{
		name:    name,
		pkgName: pkgName,
		UT:      typ,
	}
}

func (t *Named) GetPackageName() string {
	return t.pkgName
}

func (t *Named) Underlying() Type {
	if t.UT == nil {
		panic("Named type " + t.pkgName + "." + t.name + ": Underlying is nil")
	}
	return t.UT
}
func (t *Named) String() string { return t.name }
