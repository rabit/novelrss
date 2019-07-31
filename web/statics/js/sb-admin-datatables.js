// Call the dataTables jQuery plugin
$(document).ready(function() {
  $('#dataTable').DataTable({
      "iDisplayLength": 50,
  });
  $("#datatable-update-time").text("Updated " + moment().calendar());
});
