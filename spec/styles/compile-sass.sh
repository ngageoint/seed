# Compile css from sass
node-sass --include-path=styles/sass --source-map=true styles/html.scss styles/html.css
# Inject logo inline into css
sed -i.bak 's^url("images/logo.svg")^url("data:image/svg+xml;base64,'$(cat images/logo.svg|base64 -w0)'")^g' styles/html.css