const rewireReactHotLoader = require("react-app-rewire-hot-loader");

/* config-overrides.js */
module.exports = function override(config, env) {
  if (env === "development") {
    config.resolve.alias["react-dom"] = "@hot-loader/react-dom";
  }
  config = rewireReactHotLoader(config, env);
  return config;
};
