const webpack = require('webpack');
const {resolve} = require('path');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const DEVELOPMENT = 'development';

module.exports = env => {
    let appEnv = DEVELOPMENT;
    if (env && env.APP_ENV) {
        appEnv = env.APP_ENV;
    }

    let frontendVersion = '"Unknown"';
    if (env && env.VERSION) {
        frontendVersion = `"${env.VERSION}"`;
    }


    let country = 'ab';
    if (env && env.COUNTRY) {
        country = `"${env.COUNTRY}"`;
    }
    country = country.replace(/['"]/g,'')

    console.log(`Build ${appEnv} environment version ${frontendVersion} for country ${country}`);

    return {
        mode: appEnv == DEVELOPMENT ? 'development' : 'production',
        context: __dirname,
        devtool: "source-map",
        entry: ["./js-sources/index.tsx",],
        output: {
            path: __dirname + "/dist",
            filename: "wwmap.regional.js",
            publicPath: '/',
            libraryTarget: 'var',
            library: 'wwmap_regional'
        },
        resolve: {
            extensions: ['.ts', '.tsx', '.js']
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
            new webpack.DefinePlugin({
                FRONTEND_VERSION: frontendVersion,
            }),
            new HtmlWebpackPlugin({
                template: 'js-sources/index.html',
                inject: 'body',
            }),
        ],
        module: {
            rules: [
                {
                    test: /js-sources\/\.(gif|png|jpe?g|svg)$/i,
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
                    test: /js-sources\/.*?\.js$/,
                    exclude: /node_modules/,
                    loader: "babel-loader"
                },
                {
                    test: /js-sources\/.*?\.tsx?$/,
                    exclude: /node_modules/,
                    loader: "ts-loader"
                },
                {
                    test: /js-sources\/country-settings\/ab\.ts$/,
                    loader: 'file-replace-loader',
                    options: {
                        condition: 'if-replacement-exists',
                        replacement: resolve(`./js-sources/country-settings/${country}.ts`),
                        async: true,
                    }
                },
                {
                    test: /js-sources\/config\.js$/,
                    loader: 'file-replace-loader',
                    options: {
                        condition: appEnv !== DEVELOPMENT,
                        replacement: resolve('./js-sources/config.production.js'),
                        async: true,
                    }
                },
                {
                    test: /\.css$/,
                    use: ['style-loader', 'css-loader'],
                },
                {
                    test: /\.(png|jpg|gif|svg|ttf|eot|woff|woff2)$/,
                    loader: 'file-loader',
                    options: {
                        name: '[path][name].[ext]'
                    }
                },
            ]
        },
        devServer: {
            contentBase: './dist',
            compress: true,
            port: 9000,
            historyApiFallback: true
        },
    }
};