package icons

import (
	"bufio"
	"image"
	"image/png"
	"io"
	"os"

	"github.com/biessek/golang-ico"
	"github.com/develar/app-builder/pkg/fs"
	"github.com/develar/app-builder/pkg/util"
	"github.com/develar/errors"
)

const (
	PNG = 0
	ICO = 1
)

// sorted by suitability
var icnsTypesForIco = []string{ICNS_256, ICNS_256_RETINA, ICNS_512, ICNS_512_RETINA, ICNS_1024}

func LoadImage(file string) (image.Image, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer reader.Close()

	bufferedReader := bufio.NewReader(reader)

	isIcns, err := IsIcns(bufferedReader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if isIcns {
		subImageInfoList, err := ReadIcns(bufferedReader)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		for _, osType := range icnsTypesForIco {
			subImage, ok := subImageInfoList[osType]
			if ok {
				reader.Seek(int64(subImage.Offset), 0)
				bufferedReader.Reset(reader)
				// golang doesn't support JPEG2000
				return DecodeImageAndClose(bufferedReader, reader)
			}
		}

		return nil, NewImageSizeError(file, 256)
	}

	return DecodeImageAndClose(bufferedReader, reader)
}

func DecodeImageConfig(file string) (*image.Config, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result, _, err := image.DecodeConfig(reader)
	if err != nil {
		reader.Close()

		if err == image.ErrFormat {
			err = &ImageFormatError{file, "ERR_ICON_UNKNOWN_FORMAT"}
		}
		return nil, errors.WithStack(err)
	}

	err = reader.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &result, nil
}

func DecodeImageAndClose(reader io.Reader, closer io.Closer) (image.Image, error) {
	result, _, err := image.Decode(reader)
	if err != nil {
		closer.Close()
		return nil, errors.WithStack(err)
	}

	err = closer.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return result, nil
}

func SaveImage(image image.Image, outFileName string, format int) error {
	outFile, err := fs.CreateFile(outFileName)
	if err != nil {
		return errors.WithStack(err)
	}

	return SaveImage2(image, outFile, format)
}

func SaveImage2(image image.Image, outFile *os.File, format int) error {
	writer := bufio.NewWriter(outFile)

	var err error
	if format == PNG {
		err = png.Encode(writer, image)
	} else {
		err = ico.Encode(writer, image)
	}

	if err != nil {
		outFile.Close()
		return err
	}

	return util.CloseAndCheckError(writer.Flush(), outFile)
}