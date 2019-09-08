# NDR2 Ernährungsdocs Rezeptdatenbank

Dieses Repository enthält Tools um die NDR2 Ernährungsdocs Rezepte Webseite:

https://www.ndr.de/fernsehen/sendungen/die-ernaehrungsdocs/rezepte/index.html

herunter zu laden und in eine .epub Datei zu konvertieren. Diese Datei kann
dann von den meisten EBook Lesegeräten verwendet werden, für Kindle gibt es noch
im Script mkbook.sh eine auskommentierte Zeile um die .epub Datei mit Hilfe eines
Calibre tools in eine Kindle Datei zu konvertieren.

## Benötigte Software

* Go 1.13
* Pandoc
* Calibre (Optional, für Kindle Konvertierung)

Auf MacOS können die erforderlichen Pakete mit Homebrew installiert werden:

```sh
brew install golang pandoc calibre
```

In der Datei mkbook.sh sind die Schritte aufgeführt, um so ein EBook zu
erstellen. Der erste Schritt dient zum herunterladen der Webseite und
Konvertierung in das Markdown Format:

```sh
go run edocs.go
```

Das Resultat findet sich im Verzeichnis outdir.

Das Kommando zum Erstellen lautet:

```sh
pandoc -t epub -o edocs.epub --toc-depth=2 edocs.yaml --resource-path=outdir outdir/*/*.md
```

Auf einem Mac kann dann die Datei edocs.epub mit dem iBooks Programm gelesen
werden:

```sh
open edocs.epub
```

Zur Konvertierung in das Kindle format wird ein Programm von Calibre verwendet:

```sh
ebook-convert edocs.epub edocs.mobi
```
