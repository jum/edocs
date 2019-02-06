(
cd outdir
pandoc -t epub -o ../edocs.epub --toc-depth=2 ../edocs.yaml */*.md
)
#ebook-convert edocs.epub edocs.mobi
