package main

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
	"log"
	"archive/zip"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
	"sort"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("pagepal: file not specified")
	}

	gtk.Init(nil)

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("pagepal: unable to create window:", err)
	}
	win.SetTitle("PagePal")
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	imagePath := os.Args[1]
	
	dst, err := os.MkdirTemp("", "unzipped-")
    if err != nil {
        panic(err)
    }
    defer os.RemoveAll(dst)

    archive, err := zip.OpenReader(imagePath)
    if err != nil {
        panic(err)
    }
    defer archive.Close()

	var pages []string

    for _, f := range archive.File {
        filePath := filepath.Join(dst, f.Name)
        fmt.Println("unzipping file ", filePath)

        if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
            fmt.Println("invalid file path")
            return
        }
        
        if f.FileInfo().IsDir() {
            fmt.Println("creating directory...")
            os.MkdirAll(filePath, os.ModePerm)
            continue
        }

        if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
            panic(err)
        }

        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            panic(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            panic(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()

		pages = append(pages, filePath)
    }

	sort.Strings(pages)

	page := 0

	pixbuf, err := gdk.PixbufNewFromFile(pages[page])
	if err != nil {
		log.Fatal("pagepal: unable to load image:", err)
	}

	width := pixbuf.GetWidth() / 4
	height := pixbuf.GetHeight() / 4

	scaledPixbuf, err := pixbuf.ScaleSimple(width, height, gdk.INTERP_BILINEAR)
	if err != nil {
		panic(err)
	}

	image, err := gtk.ImageNewFromPixbuf(scaledPixbuf)
	if err != nil {
		panic(err)
	}
	
	win.Connect("button-press-event", func(wnd *gtk.Window, ev *gdk.Event) bool {
		btn := gdk.EventButtonNewFromEvent(ev)
		switch btn.Button() {
		case gdk.BUTTON_PRIMARY:

			winWidth, _ := wnd.GetSize()

			if int(btn.X()) > winWidth/2 {
				page++
			} else {
				page--
			}

			if page < 0 {
				page = 0
				return false
			}

			if page >= len(pages) {
				page = len(pages) - 1
				return false
			}

			newPixbuf, err := gdk.PixbufNewFromFile(pages[page])
			if err != nil {
				log.Fatal("pagepal: unable to load image:", err)
			}

			newWidth := newPixbuf.GetWidth() / 4
			newHeight := newPixbuf.GetHeight() / 4
			newScaledPixbuf, err := newPixbuf.ScaleSimple(newWidth, newHeight, gdk.INTERP_BILINEAR)
			if err != nil {
				panic(err)
			}

			win.Remove(image)

			image, err = gtk.ImageNewFromPixbuf(newScaledPixbuf)
			if err != nil {
				panic(err)
			}

			win.Add(image)
			win.ShowAll()

			win.Resize(newWidth, newHeight)
			return true
		default:
			return false
		}
	})
	
	win.Add(image)

	win.SetDefaultSize(width, height)

	win.ShowAll()

	gtk.Main()
}