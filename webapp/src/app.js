requirejs.config({
    baseUrl: 'js/bower_components',
    paths: {
        app: '../app'
    }
});

requirejs(['app/commitview']);
