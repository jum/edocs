pandoc -t epub -o edocs.epub --toc-depth=2 -N edocs.yaml --resource-path=outdir outdir/*/*.md
#ebook-convert edocs.epub edocs.mobi
