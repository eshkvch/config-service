module.exports = {
  generateKey: function (userContext, events, done) {
    userContext.vars.key = "key-" + Math.random().toString(36).substring(7);
    return done();
  }
};