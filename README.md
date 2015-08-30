A program for parsing OpenType fonts in Golang.

The main contribution of this repository is the [SFNT](https://godoc.org/github.com/ConradIrwin/font/sfnt) library which provides support for parsing OpenType, TrueType and wOFF fonts.

Also included is a sample program called `font` that will tell you information about the font file that you have.

```
go get github.com/ConradIrwin/font
font <~/Downloads/Fanwood.ttf
```

Still missing is support for parsing EOT files (which should be easy to add) and for parsing wOFF2 files (which might be more time consuming, as that uses custom compression algorithm). Also support for generating wOFF files (which is annoyingly fiddly due to the checksum calculation), and a whole load of code around dealing with the hundreds of other SFNT table formats.

Font file formats
=================

On the web there are four main types of font file, TrueType, OpenType, wOFF, and EOT. They all represent the same SFNT information, but are encoded slightly differently. You may also come across SVG fonts, which are a totally different beast.

Inside one of these files, there are two main types of glyphs, TrueType and
OpenType (also known as PostScript Type 2, or CFF). There are also a series of supporting
tables which contain meta-data about the font (its Name, Copyright Information, Kerning tweaks, Ligatures, etc.etc.)

To confuse things further, OpenType fonts use exactly the same format as TrueType fonts, and a wOFF file can contain an OpenType glyphs or a TrueType glyphs. There's no good solution to the ambiguity in terminolgy, just be aware of it.

License
=======

Copyright (c) Conrad Irwin 2015, MIT license. See LICENSE.MIT for details

