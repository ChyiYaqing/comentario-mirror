'use strict';

import gulp from 'gulp';
import dartSass from 'sass';
import gulpSass from 'gulp-sass';
import gif from 'gulp-if';
import sourcemaps from 'gulp-sourcemaps';
import cleanCss from 'gulp-clean-css';
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
    devtool: isProd ? undefined : 'eval-source-map',
}

const sources = {
    scss:       './scss/*.scss',
    typescript: './src/**/*.ts',
};

const buildDir = '../build/frontend/';

/** Lint Javascript and Typerscript sources. */
export const lint = () =>
    src(sources.typescript)
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
        .pipe(dest(buildDir));

/** Compile Typescript files. */
const compileTypescript = () =>
    src('./src/index.ts')
        .pipe(webpack(webpackConfig))
        .pipe(dest(buildDir));

/** Run all build tasks in parallel. */
export const build = parallel(compileCss, compileTypescript);

/** Lint and build all by default. */
export default series(lint, build);
