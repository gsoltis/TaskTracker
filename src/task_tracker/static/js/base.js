$(document).ready(function() {
    var auth = firebase.auth();
    var user_ = null;
    var authToken_;
    auth.onAuthStateChanged(function(user) {
        if (user) {
            console.log('got a user');
            console.log(user);
            var needsLogin = false;
            if (!user_ && !Cookies.get('_s')) {
                var cookie = Cookies.get('_s');
                console.log(cookie);
                needsLogin = true;
            }
            user_ = user;
            user_.getToken().then(function(token) {
                authToken_ = token;
                console.log(token);
                if (needsLogin) {
                    console.log('Posting now...');
                    $.ajax("/login", {
                        data: token,
                        method: 'POST'
                    }).done(function() {
                        console.log('succeeded');
                    }).fail(function(jq, status, error) {
                        console.log('Failed', status, error);
                    });
                }
                //Cookies.set('gtoken', token);
            }).catch(function(err) {
                console.error('Failed to get token');
                console.log(err);
            });
        } else {
            console.log('no user');
            showLogin();
        }
    });

    function showLogin() {
        $('#logged_out').toggle();
        $('#google-login').click(function(e) {
            e.preventDefault();
            $('.login').prop('disabled', true);
            console.log('clicked');
            var provider = new firebase.auth.GoogleAuthProvider();
            auth.signInWithPopup(provider).then(function(result) {
                console.log(result);
            }).catch(function(error) {
                console.log(error);
            });
        });
    }
    
    function showLoggedIn() {
        
    }
});
var provider = new firebase.auth.GoogleAuthProvider();