package internal

import (
	"archive/zip"
	"compress/flate"
	"io"
)

type zipWrite struct {
	writer *zip.Writer
}

// ZipWriter public interface for Zipping
type ZipWriter interface {
	Close() error
	AddFile(string, []byte) error
}

// NewZipper construct a new zip writer
func NewZipper(writer io.Writer) ZipWriter {
	zipwrite := zip.NewWriter(writer)
	zipwrite.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	return &zipWrite{writer: zipwrite}
}

func (zipper *zipWrite) Close() error {
	return zipper.writer.Close()
}

func (zipper *zipWrite) AddFile(name string, data []byte) error {
	f, err := zipper.writer.CreateHeader(&zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	})
	if err != nil {
		return err
	}
	_, err = f.Write(data)

	return err
}
