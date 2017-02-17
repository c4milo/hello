const ExtractTextPlugin = require('extract-text-webpack-plugin');
const webpack           = require('webpack');
var autoprefixer        = require('autoprefixer');

module.exports = {
	entry: {
		app: "./app/index.js",
		vendor: []
	},
	output: {
		path: __dirname + '/public',
		filename: '[name].js',
		chunkFilename: '[id].js',
		// Webpack dev server keeps compiled assets in memory and serve
		// them once an HTTP request is made to publicPath
		publicPath: '/'
	},

	module: {
		loaders: [{
			test: /\.js$/,
			exclude: /node_modules/,
			loader: "babel",
			include: __dirname
		}, {
			test: /\.scss$/,
			loader: ExtractTextPlugin.extract("style-loader", "css-loader!sass-loader!postcss-loader")
		}]
	},
	postcss: function () {
		return [autoprefixer];
	},
	plugins: [
		new ExtractTextPlugin("[name].css"),
		// Removes vendored modules from app.js
		new webpack.optimize.CommonsChunkPlugin(/* chunkName= */"vendor", /* filename= */"vendor.js"),
		// Allows replacing variables based on the running environment
		new webpack.DefinePlugin({
			'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development')
		}),
	],

	devServer: {
		proxy: [{
			path: '*',
			target: 'https://localhost:9999',
			bypass: function(req, res, proxyOptions) {
				if (req.headers.accept.indexOf('html') !== -1) {
					console.log('Skipping proxy for browser request.');
					return '/index.html';
				}
			}
		}]
	}
};
