# Compile css from sass
node-sass --include-path=styles/sass --source-map=true styles/html.scss styles/html.css
# Inject logo inline into css - busybox doesn't support base64 -w0 flag, engage linux foo
sed -i.bak 's^url("images/logo.svg")^url("data:image/svg+xml;base64,'$(cat images/logo.svg|base64|xargs|sed -e 's/ //g')'")^g' styles/html.css
