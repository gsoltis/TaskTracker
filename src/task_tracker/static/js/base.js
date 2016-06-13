$(document).ready(function() {
    var auth = firebase.auth();
    var user_ = null;
    var authToken_;
    auth.onAuthStateChanged(function(user) {
        if (user) {
            var needsLogin = false;
            if (!user_ && !Cookies.get('_s')) {
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
                        showLoggedIn();
                    }).fail(function(jq, status, error) {
                        console.log('Failed', status, error);
                    });
                } else {
                    showLoggedIn();
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
        $('#logged_out').show();
        $('#logged_in').hide();
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
        $('#logged_out').hide();
        $('#logged_in').show();
        $('#add-task').prop('disabled', true);
        $.ajax("/api/tasks").done(function(tasks) {
            console.log("tasks", tasks);
            tasks_ = tasks;
            renderTasks();
            $('#add-task').prop('disabled', false);
        }).fail(function(jq, status, error) {
            console.log('Failed to get tasks', status, error);
        });
    }

    var tasks_ = {};

    var addTask = function(task_key, task) {
        var li = $('<li class="task" id="' + task_key + '">' + task['Name'] + '</li>');
        $('#task-list').append(li);
    };

    function renderTasks() {
        for (var task_key in tasks_) {
            if (tasks_.hasOwnProperty(task_key)) {
                addTask(task_key, tasks_[task_key]);
            }
        }
    }

    $('#task-submit').click(function(e) {
        e.preventDefault();
        $('#add-task').prop('disabled', true);
        var task_name = $('#task-name').val();
        if (!task_name) {
            return;
        }
        $.ajax("/api/tasks", {
            method: 'POST',
            data: task_name
        }).done(function(task_key) {
            console.log('Set task: ' + task_key);
            addTask(task_key, { Name: task_name });
            $('#task-name').val('');
            $('#add-task').prop('disabled', false);
        }).fail(function(jq, status, error) {
            console.log("failed", status, error);
        });
    })
});