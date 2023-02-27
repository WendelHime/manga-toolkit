# mangaimgstopdf
This repo was created with the intuiton of downloading ZIP images from mangafreak and merge them into a PDF with pages A5, so I can read them on any device (like kindle, remarkable, onyx)

# Cloning the repo
```bash
git clone git@github.com:WendelHime/manga-toolkit.git
```

# Building executables
Change directory to the cloned repo
```bash
cd manga-toolkit
```

Build downloader
```bash
go build ./cmd/download_zips
```

Build PDF generator
```bash
go build ./cmd/generate_pdf_from_zip
```

# Running

```bash
$ ./download_zips -h                                                                                                      
Usage of ./download_zips:
This script download the manga ZIPs from mangafreak  -from_chapter int
    	From which chapter should be downloaded (default 1)
  -manga_term string
    	Manga term used on URL for downloading ZIP files (default "Chainsaw_Man")
  -output_dir string
    	output directory that will contain the ZIP files (default "/Users/wotan/Documents/ChainsawMan")
  -to_chapter int
    	To which chapter to be downloaded (default 120)
```

```bash
$ ./generate_pdf_from_zip -h                                                                                              
Usage of ./generate_pdf_from_zip:
This script generates PDFs from a provided directory containing ZIP files with JPG images.
/home/someuser/mangadir
| chapter1.zip
  - img1.jpg
  - img2.jpg
| chapter2.zip
| chapter3.zip
  -input_dir string
    	directory with manga ZIP images
  -output_dir string
    	output directory that will contain generated PDFs
```
