package work

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProcessFile ... process the downloaded file
func ProcessFile(file []byte, src Source) error {

	log.Println("Process file")
	folders := strings.Split(src.GetPath(), "/")
	i := len(folders) - 1
	a := strings.Join(folders[i:], "/")
	fullPath := TmpDir + "/" + a
	log.Println("Copying file: ", fullPath, a)
	err := ioutil.WriteFile(fullPath, file, 0x777)
	if err != nil {
		return err
	}

	//switch artifactType {
	//case "", "archive":
		// Unzip
		log.Println("Unzipping file: ", fullPath)
		err = unzip(fullPath, TmpDir)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to unzip file: %v %v", fullPath, err.Error()))
		}

		log.Println("Removing file: ", fullPath)
		err = os.Remove(fullPath)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not remove zipfile: %v %v", fullPath, err.Error()))
		}

		err = GetDirectory(src)
		if err != nil {
			log.Println("ORGANIZE_ERROR: ", err.Error())
			return err
		}
	//}

	return nil

}

func GetDirectory(src Source) error {

	baseDirHandle, err := os.Open(TmpDir)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not open directory: %v", err.Error()))
	}
	defer baseDirHandle.Close()
	baseDirContents, err := baseDirHandle.Readdir(0)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not open directory: %v", err.Error()))
	}
	// Only one directory is created
	exp := regexp.MustCompile("^[a-zA-Z0-9].*$")
	for _, baseDirObject := range baseDirContents {
		if baseDirObject.IsDir() && exp.MatchString(baseDirObject.Name()){
			src.SetWorkDir(TmpDir + "/" + baseDirObject.Name())
		}
	}
	return nil
}

/*
func PackageFiles(builder *propeller.Builder) (*bytes.Buffer, error) {

	log.Println("Packaging files for build: ", builder.Name)
	buf := new(bytes.Buffer)

	// tar write
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := IterDirectory(artifactRoot, tw)
	if err != nil {
		return buf, errors.New(err.Error())
	}

	log.Println("Packaging complete: ", builder.Name)
	return buf, nil
}
*/

/*
// IterDirectory goes through each directory
func IterDirectory(dirPath string, tw *tar.Writer) error {

	dir, err := os.Open(dirPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Tar failure opening directory: %v", err.Error()))
	}

	defer dir.Close()
	fis, err := dir.Readdir(0)
	if err != nil {
		return errors.New(fmt.Sprintf("Tar failure reading directory: %v", err.Error()))
	}

	for _, fi := range fis {
		curPath := dirPath + "/" + fi.Name()
		log.Println("CURPATH: ", curPath)
		if fi.IsDir() {
			log.Println("ISDIR")
			//TarGzWrite( curPath, tw, fi )
			err = IterDirectory(curPath, tw)
			if err != nil {
				return errors.New(err.Error())
			}

		} else {
			log.Printf("adding... %s\n", curPath)
			TarGzWrite(curPath, tw, fi)
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}

	//time.Sleep(10 * time.Second)
	return nil
}
*/

/*
func TarGzWrite(_path string, tw *tar.Writer, fi os.FileInfo) error {

	log.Println("TAR_GZ_WRITE_PATH: ", _path)
	fr, err := os.Open(_path)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to open: %v", err.Error()))
	}
	defer fr.Close()
	h := new(tar.Header)
	h.Name = _path
	h.Size = fi.Size()
	h.Mode = int64( fi.Mode())
	h.ModTime = fi.ModTime()
	e := ""

	err = tw.WriteHeader(h)
	if err != nil {
		e = fmt.Sprintf("Tar failure during header write: %v", err.Error())
		log.Println("XXX_Error")
		return errors.New(e)
	}

	_, err = io.Copy(tw, fr)
	if err != nil {
		return errors.New(fmt.Sprintf("Tar failure during copy: %v", err.Error()))
	}

	//time.Sleep(10 * time.Second)
	return nil
}
*/

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return errors.New(fmt.Sprintf("Unzip error opening zip file: %v", err.Error()))

	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return errors.New(fmt.Sprintf("Unzip error creating a directory: %v", err.Error()))
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return errors.New(fmt.Sprintf("Unzip error extract/write: %v", err.Error()))
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return errors.New(fmt.Sprintf("Unzip error opening file: %v", err.Error()))
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return errors.New(fmt.Sprintf("Unzip error copying file: %v", err.Error()))
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return errors.New(fmt.Sprintf("Unzip error extract and write: %v", err.Error()))
		}
	}
	return nil
}

/*
func ProcessFileStream(body io.ReadCloser, artifact *propeller.Artifact) error {
	folders := strings.Split(artifact.ArtifactName, "/")
	i := len(folders) - 1
	a := strings.Join(folders[i:], "/")
	fullPath := tmpdir + "/" + a
	log.Println("Downloading file from stream to:", fullPath)
	out, err := os.Create(fullPath)
	defer out.Close()
	if err != nil {
		log.Println("Create Error:", err.Error())
		return err
	}
	// copy file
	n, err := io.Copy(out, body)
	if err != nil {
		log.Println("Copy Error:", err.Error())
		return err
	}
	log.Println("Downloaded, number of bytes written: ", n)
	err = os.Chmod(fullPath, 0x777)
	if err != nil {
		log.Printf("Unable to set permissions on file: %v %v", fullPath, err.Error())
		// don't return, we may still succeed
	}

	_, err = ioutil.ReadFile(fullPath)
	if err != nil {
		log.Println("ReadFile Error:", err.Error())
		return errors.New(fmt.Sprintf("Unable to read file: %v %v", fullPath, err.Error()))
	}

	switch artifact.FileType {
	case "", "archive":
		// Unzip
		log.Println("Unzipping file: ", fullPath)
		err = unzip(fullPath, tmpdir)
		if err != nil {
			return errors.New(fmt.Sprintf("Unable to unzip file: %v %v", fullPath, err.Error()))
		}

		log.Println("Removing file: ", fullPath)
		err = os.Remove(fullPath)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not remove zipfile: %v %v", fullPath, err.Error()))
		}
		GetDirectory(artifact)
	}
	return nil
//}
*/