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
                if (needsLogin) {
                    $.ajax("/login", {
                        data: token,
                        method: 'POST'
                    }).done(function() {
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
            showLogin();
        }
    });

    function showLogin() {
        $('#logged_out').show();
        $('#logged_in').hide();
        $('#google-login').click(function(e) {
            e.preventDefault();
            $('.login').prop('disabled', true);
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
        disableTasks();
        disableGoals();
        $.ajax("/api/tasks").done(function(tasks) {
            tasks_ = tasks;
            renderTasks();
            enableTasks();
        }).fail(function(jq, status, error) {
            console.log('Failed to get tasks', status, error);
        });
        $.ajax("/api/goals").done(function(goals) {
            goals_ = goals;
            renderGoals();
            enableGoals();
        }).fail(function(jq, status, error) {
            console.log('Failed to get goals', status, error);
        });
    }

    var tasks_ = {};
    var goals_ = {};

    function enableTasks() {
        $('#add-task').prop('disabled', false);
    }

    function disableTasks() {
        $('#add-task').prop('disabled', true);
    }

    function disableGoals() {
        $('#add-goal').prop('disabled', true);
    }

    function enableGoals() {
        $('#add-goal').prop('disabled', false);
    }

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

    function renderGoals() {
        // TODO: implement this
    }

    $('#goal-submit').click(function(e) {
        e.preventDefault();
        disableGoals();
        var task_id = '5768037999312896';
        var numerator = 3;
        var denominator = "week";
        var goal = {
            task_id: task_id,
            numerator: numerator,
            denominator: denominator
        };
        $.ajax("/api/goals", {
            method: 'POST',
            data: JSON.stringify(goal)
        }).done(function(goal_id) {
            console.log('Set goal: ' + goal_id);
            addGoal(goal_id, goal);
            // TODO: reset form
            enableGoals();
        });
    });

    $('#task-submit').click(function(e) {
        e.preventDefault();
        disableTasks();
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
            enableTasks();
        }).fail(function(jq, status, error) {
            console.log("failed", status, error);
        });
    })
});