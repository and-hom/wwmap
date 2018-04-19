var wwmap = angular.module("wwmap", ["xeditable"]);

wwmap.run(function(editableOptions) {
  editableOptions.theme = 'bs3';
});

wwmap.controller('Ctrl', function($scope, $filter, $http, $q) {
  $scope.points = [];

  $scope.categories = [
    {value: '1', text: '1'},
    {value: '2', text: '2'},
    {value: '3', text: '3'},
    {value: '4', text: '4'},
    {value: '5', text: '5'},
    {value: '6', text: '6'}
  ];

  $scope.rivers = {};
  $scope.loadRivers = function(point) {
    url = apiBase + '/nearest-rivers?lat=' + point.lat + '&lon=' + point.lon
    return $http.get(url).success(function(data) {
      data.push({"id":null,"title":"Нет в этом списке"})
      $scope.setRivers(point, data)
    });
  };
  $scope.setRivers = function(point,rivers) {
    $scope.rivers[point.lat+'-'+point.lon] = rivers
  }

  $scope.getRivers = function(point) {
    return $scope.rivers[point.lat+'-'+point.lon]
  }

  $scope.showRiver = function(point) {
    if (!point.waterway_id || point.waterway_id==0) {
      return null
    }
    selected = $filter('filter')($scope.getRivers(point), {id: point.waterway_id})
    if (selected.length == 0) {
        return null
    }
    var river_label = selected[0].title
    if (selected[0].osm_id) river_label +=' (' + selected[0].osm_id + ')'
    return river_label
  };

  $scope.showCategory = function(point) {
    var selected = [];
    if(point.category) {
      selected = $filter('filter')($scope.categories, {value: point.category});
    }
    return selected.length ? selected[0].text : 'Not set';
  };

  $scope.checkTitle = function(data) {
    if (!data) {
      return "Title should not be empty";
    }
  };


  $scope.checkLength = function(data, maxLen) {
    if (data && data.length > maxLen) {
      return "Text is too long";
    }
  };

  $scope.checkNumber = function(data, allowEmpty) {
    if ((data && isNaN(data)) || (!allowEmpty && !data)) {
      return "Not a number!";
    }
  };

  // filter points to show
  $scope.filterPoints = function(point) {
    return point.isDeleted !== true;
  };

  // mark point as deleted
  $scope.deletePoint = function(title) {
    var filtered = $filter('filter')($scope.points, {title: title});
    if (filtered.length) {
      filtered[0].isDeleted = true;
    }
  };

  // add point
  $scope.addPoint = function() {
    $scope.points.push({
      id: $scope.points.length+1,
      name: '',
      status: null,
      group: null,
      isNew: true
    });
  };

  // cancel all changes
  $scope.cancel = function() {
    for (var i = $scope.points.length; i--;) {
      var point = $scope.points[i];
      // undelete
      if (point.isDeleted) {
        delete point.isDeleted;
      }
      // remove new
      if (point.isNew) {
        $scope.points.splice(i, 1);
      }
    };
  };


  $scope.append = function(points) {
    $scope.points.push(...points)
  }


  // save edits
  $scope.saveTable = function() {
    var results = [];
    for (var i = $scope.points.length; i--;) {
      var point = $scope.points[i];
      // actually delete point
      if (point.isDeleted) {
        $scope.points.splice(i, 1);
      }
      // mark as not new
      if (point.isNew) {
        point.isNew = false;
      }

      results.push(point);
    }

    dataForBackend = results.map(function (v) {
            v.point = [parseFloat(v.lat), parseFloat(v.lon)]
            return v
       })
    $http({
        method: 'POST',
        url:apiBase + '/whitewater',
        data:dataForBackend})
    .then(function successCallback(response) {
        window.location.href="./map.htm"
    }, function errorCallback(response) {
        window.alert("Publish failed: " + response.statusText)
    });

    return $q.all(results);
  };
});