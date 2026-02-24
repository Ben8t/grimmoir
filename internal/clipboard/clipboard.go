package clipboard

import atotto "github.com/atotto/clipboard"

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) ReadAll() (string, error) {
	return atotto.ReadAll()
}

func (a *Adapter) WriteAll(text string) error {
	return atotto.WriteAll(text)
}
