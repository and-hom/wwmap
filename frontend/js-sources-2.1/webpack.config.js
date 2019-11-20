const webpack = require('webpack');
const {resolve} = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');

const DEVELOPMENT = 'development';

module.exports = env => {
    let appEnv = DEVELOPMENT;
    if (env && env.APP_ENV) {
        appEnv = env.APP_ENV;
    }

    console.log(`Build ${appEnv} environment`);

    return {
        mode: appEnv == DEVELOPMENT ? 'development' : 'production',
        context: __dirname,
        devtool: "source-map",
        entry: ["./main.js",],
        output: {
            path: __dirname + "/../js",
            filename: "map.v2.1.js",
            publicPath: './js/',
            libraryTarget: 'var',
            library: 'wwmap'
        },
        plugins: [
            new webpack.ProvidePlugin({
                $: 'jquery',
                jQuery: 'jquery',
                'window.jQuery': 'jquery',
            }),
            new MiniCssExtractPlugin({
                filename: '[name].css',
                chunkFilename: '[id].css',
                ignoreOrder: false,
            }),
        ],
        module: {
            rules: [
                {
                    test: /.*?\.js$/,
                    exclude: /node_modules/,
                    loader: "babel-loader"
                },
                {
                    test: /config\.js$/,
                    loader: 'file-replace-loader',
                    options: {
                        condition: appEnv !== DEVELOPMENT,
                        replacement: resolve('./config.production.js'),
                        async: true,
                    }
                },
                {
                    test: /\.css$/,
                    use: ['style-loader', 'css-loader'],
                },
                {
                    test: /\.(png|jpg|svg|ttf|eot|woff|woff2)$/,
                    loader: 'file-loader',
                    options: {
                        name: '[path][name].[ext]'
                    }
                }
            ]
        },
    }
};