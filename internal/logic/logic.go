// Package logic contains the logic used on manga-toolkit clients
package logic

import (
	"archive/zip"
	"io"
)

// Logic abstracts the logic used on manga-toolkit
type Logic interface {
	GeneratePDFFromZip(reader *zip.ReadCloser, output io.WriteCloser) error
}

type logic struct{}

// NewLogic builds the logic used on manga-toolkit
func NewLogic() Logic {
	return logic{}
}
