echo Patching doc for PDF generation...
apk -U add perl
# perl -0777 -i.bak -pe 's/\/\/\{pdf\+4}\na/4\+a/g;' -pe 's/\/\/\{pdf\+1}\n3\+a/4\+a/g' /documents/sections/standard.adoc

echo Generating PDF...
# asciidoctor-pdf -a pdf-style=styles/pdf-theme.yml -a pdf-fontsdir=styles/fonts/ -D /documents/output index.adoc
asciidoctor-pdf -D /documents/output index.adoc

# echo Replacing original doc following PDF generation...
# mv /documents/sections/standard.adoc.bak /documents/sections/standard.adoc