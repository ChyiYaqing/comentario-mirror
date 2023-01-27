const path = require('path');

module.exports = {
    entry: './frontend/src/index.ts',
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
        ],
    },
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
    },
    output: {
        filename: 'commento.js',
        path: path.resolve(__dirname, 'build'),
    },
};
