To rebuild the JS, create the minified file and the corresponding map file.

```
npm install

npm run build
```

There are convenience functions to help during development. First replace `bundle.min.js` in `tmpl/index.html` by `bundle.js` and run 

```
npm run watch-dev
```
 
It will rebuild the bundle whenever a js file has been changed.
