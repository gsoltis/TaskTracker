$(document).ready(function() {
  var auth = firebase.auth();
  var user_ = null;
  var authToken_;
  auth.onAuthStateChanged(function(user) {
    if (user) {
      var needsLogin = false;
      var sessionCookie = Cookies.get('_s');
      console.log('session cookie: ', sessionCookie);
      if (!user_ && !sessionCookie) {
        needsLogin = true;
      }
      user_ = user;
      user_.getToken().then(function(token) {
        authToken_ = token;
        if (needsLogin) {
          console.log('needs login');
          $.ajax("/login", {
            data: token,
            method: 'POST'
          }).done(function() {
            showLoggedIn();
          }).fail(function(jq, status, error) {
            console.log('Failed', status, error);
          });
        } else {
          console.log("Doesn't need login");
          showLoggedIn();
        }
        //Cookies.set('gtoken', token);
      }).catch(function(err) {
        console.error('Failed to get token');
        console.log(err);
      });
    } else {
      user_ = null;
      showLogin();
    }
  });

  function showLogin() {
    $('#logged_out').show();
    $('#logged_in').hide();
    $('.login').prop('disabled', false);
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

    $('#logout').click(function(e) {
      e.preventDefault();
      console.log('Logging out');
      logout();
    });
  }

  var tasks_ = {};
  var goals_ = {};

  function logout() {
    tasks_ = {};
    goals_ = {};
    $('#goal-list').empty();
    $('#task-list').empty();
    auth.signOut();
    Cookies.remove('_s');
    //showLogin();
  }

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
    var option = $('<option value="' + task_key + '">' + task['Name'] + '</option>');
    $('#task-select').append(option);
  };

  function renderTasks() {
    for (var task_key in tasks_) {
      if (tasks_.hasOwnProperty(task_key)) {
        addTask(task_key, tasks_[task_key]);
      }
    }
  }

  var Periods = [
    'day',
    'week'
  ];
  function periodFromInt(i) {
    return i < Periods.length ? Periods[i] : '???';
  }

  function intFromPeriod(period) {
    return Periods.indexOf(period);
  }

  function goalLI(goal_key, goal) {
    var period = goal['Period'];
    var period_string = periodFromInt(period);
    var frequency = goal['Frequency'];
    var progress = goal['Times'];
    var progress_count = progress.length;
    var task_key = goal['TaskId'];
    var task_name = goal['Task']['Name'];
    var task_span = $('<span class="goal-task" id="' + task_key + '">' + task_name + '</span>');
    var goal_span = $('<span class="goal" id="' + goal_key + '">' + frequency + ' times every ' + period_string + '</span>');
    var percent_progress = Math.floor(((progress_count / frequency) * 100) + 0.5);
    var progress_span = $('<span>' + percent_progress + '% so far</span>');
    var btn = $('<button class="progress">Report progress</button>');
    var total = goal['Aggregations'].length;
    var completed = 0;
    for (var agg in goal['Aggregations']) {
      if (agg['Success']) {
        completed++;
      }
    }
    var aggs = $('<span>' + completed + ' out of last ' + total + '</span>');
    var li = $('<li class="goal" id="' + goal_key + '"></li>');
    li.append(task_span);
    li.append(goal_span);
    li.append(progress_span);
    li.append(aggs);
    li.append(btn);
    return li;
  }

  function addGoal(goal_key, goal) {
    var li = goalLI(goal_key, goal);
    $('#goal-list').append(li);
  }

  function refreshGoal(goal_key, goal) {
    var existing = $('#' + goal_key);
    var replacement = goalLI(goal_key, goal);
    existing.replaceWith(replacement);
  }

  function renderGoals() {
    for (var goal_key in goals_) {
      if (goals_.hasOwnProperty(goal_key)) {
        addGoal(goal_key, goals_[goal_key]);
      }
    }
  }

  $('#goal-submit').click(function(e) {
    e.preventDefault();
    disableGoals();
    var goal_root = $('#add-goal');
    var task_id = $('#task-select').val();
    var numerator = parseInt(goal_root.find('#goal-frequency').val(), 10);
    var denominator = goal_root.find('#period').val();
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
      var goal = {
        Frequency: numerator,
        Period: intFromPeriod(denominator),
        TaskId: task_id,
        Task: {
          Name: tasks_[task_id]['Name']
        },
        Times: [],
        Aggregations: []
      };
      goals_[goal_id] = goal;
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
      var task = { Name: task_name };
      tasks_[task_key] = task;
      addTask(task_key, task);
      $('#task-name').val('');
      enableTasks();
    }).fail(function(jq, status, error) {
      console.log("failed", status, error);
    });
  });

  $('#goal-list').on('click', 'button.progress', function(e) {
    e.preventDefault();
    var btn = $(e.target);
    btn.prop('disabled', true);
    var goal_id = $(this).parent().attr('id');
    var epoch = Math.floor(new Date().getTime() / 1000);
    console.log('Adding progress for ' + goal_id + '!');
    $.ajax('/api/progress', {
      method: 'POST',
      data: JSON.stringify({
        goal_id: goal_id,
        epoch: epoch
      })
    }).done(function() {
      console.log('Recorded progress');
      goals_[goal_id]['Times'].push(epoch);
      //btn.prop('disabled', false);
      refreshGoal(goal_id, goals_[goal_id]);
    }).fail(function(jq, status, error) {
      console.log('Failed to record progress', status, error);
      btn.prop('disabled', false);
    });
  });
});