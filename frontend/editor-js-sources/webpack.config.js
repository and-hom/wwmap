const webpack = require('webpack');
const {resolve} = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const VueLoaderPlugin = require('vue-loader/lib/plugin');

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
            filename: "editor.v2.js",
            publicPath: './js/',
            libraryTarget: 'var',
            library: 'wwmap_editor'
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
            new VueLoaderPlugin(),
        ],
        module: {
            rules: [
                {
                    test: /\.(gif|png|jpe?g|svg)$/i,
                    use: [
                        'file-loader',
                        {
                            loader: 'image-webpack-loader',
                            options: {
                                bypassOnDebug: true, // webpack@1.x
                                disable: true, // webpack@2.x and newer
                            },
                        },
                    ],
                },
                {
                    test: /\.vue$/,
                    loader: 'vue-loader'
                },
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
                },
            ]
        },
    }
};