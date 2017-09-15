const path = require('path');
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextPlugin = require('extract-text-webpack-plugin');
const merge = require('webpack-merge');

const env = process.env.NODE_ENV || 'dev';
const prod = env === 'prod';

const publicPath = '/';
const entryPath = './src/main.ts';

const commonConfig = {
    resolve: {
        extensions: ['.js', '.ts', '.elm', '.css', '.scss']
    },
    context: __dirname,
    plugins: [
        new webpack.HotModuleReplacementPlugin(),
        new webpack.NamedModulesPlugin(),
        new HtmlWebpackPlugin({
            template: 'src/index.html',
            inject: 'body',
            filename: 'index.html'
        })
    ],
    devServer: {
        stats: 'errors-only'
    },
    module: {
        rules: [
            {
                test: /\.scss$/,
                use: [
                    "style-loader",
                    "css-loader",
                    "sass-loader"
                ]
            },
            {
                test: /\.css$/,
                use: [
                    'style-loader',
                    'css-loader?modules'
                ],
            },
            {
                test: /\.ts$/,
                use: [
                    'awesome-typescript-loader'
                ]
            }
        ],
        noParse: /\.elm$/
    }
};

let config;

if (prod) {

    config = merge(commonConfig, {
        entry: entryPath,
        module: {
            rules: [
                {
                    test: /\.elm$/,
                    exclude: [/elm-stuff/, /node_modules/],
                    use: 'elm-webpack-loader?pathToMake=./node_modules/.bin/elm-make'
                },
                {
                    test: /\.sc?ss$/,
                    use: ExtractTextPlugin.extract({
                        fallback: 'style-loader',
                        use: ['css-loader', 'sass-loader']
                    })
                }
            ]
        },
        output: {
            path: path.resolve(__dirname, './dist'),
            filename: 'static/js/[name].bundle.js'
        },
        plugins: [
            new ExtractTextPlugin({
                filename: 'static/css/[name].css',
                allChunks: true,
            }),
            new webpack.optimize.UglifyJsPlugin()
        ]
    });

} else {

    config = merge(commonConfig, {
        entry: {
            app: [
                `webpack-dev-server/client?http://localhost:4200`,
                'webpack/hot/only-dev-server',
                entryPath
            ]
        },
        output: {
            path: path.resolve(__dirname, './dist'),
            filename: '[name].bundle.js',
            publicPath
        },
        module: {
            rules: [
                {
                    test: /\.elm$/,
                    exclude: [/elm-stuff/, /node_modules/],
                    loader: 'elm-hot-loader!elm-webpack-loader?verbose=true&warn=true&debug=true&pathToMake=./node_modules/.bin/elm-make'
                }
            ]
        }
    });

}

module.exports = config;
