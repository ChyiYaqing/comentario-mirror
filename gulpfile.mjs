'use strict';

import gulp from 'gulp';
import { deleteAsync } from 'del';
import dartSass from 'sass';
import gulpSass from 'gulp-sass';
import stripDebug from'gulp-strip-debug';
import gif from 'gulp-if';
import sourcemaps from 'gulp-sourcemaps';
import cleanCss from 'gulp-clean-css';
import htmlMinifier from 'gulp-html-minifier';
import terser from 'gulp-terser';
import concat from 'gulp-concat';
import rename from 'gulp-rename';
import eslint from 'gulp-eslint';
import webpack from 'webpack-stream';

const sass = gulpSass(dartSass);
const { dest, parallel, series, src } = gulp;

/** Whether we're running in the production mode (default). */
const isProd = ((process.env.NODE_ENV || 'production').trim().toLowerCase() === 'production');

// Load and tweak Webpack config
import webpackCfgJs from './webpack.config.js';
const webpackConfig = {
    ...webpackCfgJs,
    mode: isProd ? 'production' : 'development',
    devtool: isProd ? undefined : 'inline-source-map',
}

const sources = {
    fonts:      'frontend/fonts/**/*',
    html:       'frontend/*.html',
    images:     'frontend/images/**/*',
    javascript: 'frontend/js/**/*.js',
    scss:       'frontend/scss/*.scss',
    typescript: 'frontend/src/**/*.ts',
};

const dir = {
    css:    'build/css/',
    fonts:  'build/fonts/',
    html:   'build/html',
    images: 'build/images/',
    js:     'build/js/',
};

const jsCompileMap = {
    'jquery.js': [
        'node_modules/jquery/dist/jquery.min.js',
    ],
    'vue.js': [
        'node_modules/vue/dist/vue.min.js',
    ],
    'highlight.js': [
        'node_modules/highlightjs/highlight.pack.min.js',
    ],
    'chartist.js': [
        'node_modules/chartist/dist/chartist.min.js',
    ],
    'login.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/auth-common.js',
        'frontend/js/login.js',
    ],
    'forgot.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/forgot.js',
    ],
    'reset.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/reset.js',
    ],
    'signup.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/auth-common.js',
        'frontend/js/signup.js',
    ],
    'dashboard.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/errors.js',
        'frontend/js/self.js',
        'frontend/js/dashboard.js',
        'frontend/js/dashboard-setting.js',
        'frontend/js/dashboard-domain.js',
        'frontend/js/dashboard-installation.js',
        'frontend/js/dashboard-general.js',
        'frontend/js/dashboard-moderation.js',
        'frontend/js/dashboard-statistics.js',
        'frontend/js/dashboard-import.js',
        'frontend/js/dashboard-danger.js',
        'frontend/js/dashboard-export.js',
    ],
    'settings.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/errors.js',
        'frontend/js/self.js',
        'frontend/js/settings.js',
    ],
    'logout.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/logout.js',
    ],
    'count.js': [
        'frontend/js/count.js',
    ],
    'unsubscribe.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/unsubscribe.js',
    ],
    'profile.js': [
        'frontend/js/constants.js',
        'frontend/js/utils.js',
        'frontend/js/http.js',
        'frontend/js/profile.js',
    ],
};

/** Remove all target directories. */
export const clean = () => deleteAsync(Object.values(dir));

/** Lint Javascript and Typerscript sources. */
export const lint = () =>
    src([sources.javascript, sources.typescript])
        .pipe(eslint())
        .pipe(eslint.format())
        .pipe(eslint.failAfterError());

/** Compile SCSS files into CSS. */
const compileCss = () =>
    src(sources.scss)
        .pipe(gif(!isProd, sourcemaps.init()))
        .pipe(sass({outputStyle: isProd ? 'compressed' : 'expanded'}).on('error', sass.logError))
        // Write out source maps in non-prod mode
        .pipe(gif(!isProd, sourcemaps.write()))
        // Cleanup CSS in prod mode
        .pipe(gif(isProd, cleanCss()))
        .pipe(dest(dir.css));

/** Compile Javascript files. */
const compileJavascript = parallel(
    Object.entries(jsCompileMap)
        .map(([out, files]) => {
            const task = () => src(files)
                .pipe(gif(!isProd, sourcemaps.init()))
                .pipe(concat(out))
                .pipe(rename(out))
                // Write out source maps in non-prod mode
                .pipe(gif(!isProd, sourcemaps.write()))
                // Strip condole.log and debug statements in prod mode
                .pipe(gif(isProd, stripDebug()))
                // Minify and uglify the code in prod mode
                .pipe(gif(isProd, terser()))
                .pipe(dest(dir.js));
            task.displayName = `compile-${out}`;
            return task;
        }));

/** Compile Typescript files. */
const compileTypescript = () =>
    src('frontend/src/index.ts')
        .pipe(webpack(webpackConfig))
        .pipe(dest(dir.js));

/** Copy font files. */
const copyFonts = () => src(sources.fonts).pipe(dest(dir.fonts));

/** Copy HTML files. */
const copyHtml = () =>
    src(sources.html)
        // Minify HTML in prod mode
        .pipe(gif(isProd, htmlMinifier({collapseWhitespace: true, removeComments: true})))
        .pipe(dest(dir.html));

/** Copy image files. */
const copyImages = () => src(sources.images).pipe(dest(dir.images));

/** Run all build tasks in parallel. */
export const build = parallel(compileCss, compileJavascript, compileTypescript, copyFonts, copyHtml, copyImages);

/** Clean build all by default. */
export default series(lint, clean, build);
