var Toast;
var Table;

$(function () {
    Toast = Swal.mixin({
        toast: true,
        position: 'top-end',
        showConfirmButton: false,
        timer: 3000
    });

    var buttons =  [
        {
            text: '<i class="fas fa-eye"></i>',
            className: "show-btn",
            attr:  {
                "data-toggle": 'modal',
                "data-target": '#modal-show'
            },
            action: function () {
                var data = Table.rows( { selected: true } ).data().toArray();
                ajaxDetailData(data[0].id);
            }
        },
        {
            text: '<i class="fas fa-trash"></i>',
            className: "remove-btn",
            attr:  {
                "data-toggle": 'modal',
                "data-target": '#modal-delete'
            },
            action: function () {
                var data = Table.rows( { selected: true } ).data().toArray();
                $('#question-confirm').text(data[0].question);
                $('input[name=confirm-id]').val(data[0].id);
            }
        }
    ];
    buttons.push('pageLength');

    Table = $('#questions').DataTable({
        "dom": 'Bfrtip',
        "processing": true,
        "serverSide": true,
        "ajax": {
            "url": "/api/admin/quiz"
        },
        lengthMenu: [
            [ 5, 10, 25 ],
            [ '5 rows', '10 rows', '25 rows' ]
        ],
        "columns": [
            { "data": "checkbox" },
            { "data": "no" },
            { "data": "id" },
            { "data": "question" },
            { "data": "status" }
        ],
        "columnDefs": [ {
            "orderable": false,
            "className": 'select-checkbox',
            "targets":   0
        } ],
        "select": {
            "style":    'os',
            "selector": 'td:first-child'
        },
        "order": [[ 1, 'asc' ]],
        "buttons": buttons
    });

    Table.on( 'select', function ( e, dt, type, indexes ) {
        handleAvailableButton();
    });

    Table.on( 'deselect', function ( e, dt, type, indexes ) {
        handleAvailableButton();
    } );

    // override button style
    setDataTableButtonStyles();
    setFormUpdateListener();
    setAddForm();
    setConfirmForm();
    handleAvailableButton();

    $('#submit-question-add').prop('disabled', true);
    $('#add-question-text').keyup(function (e) {
        var value = $('#add-question-text').val();
        if(value.length > 0) {
            $('#submit-question-add').prop('disabled', false);
        }else{
            $('#submit-question-add').prop('disabled', true);
        }
    });
    $('#modal-show').on('hidden.bs.modal', function () {
        Table.ajax.reload();
    });
});

function handleAvailableButton(){
    var countSelected = Table.rows( '.selected' ).count();
    if(countSelected > 0) {
        $('.show-btn').prop('disabled', false);
        $('.remove-btn').prop('disabled', false);
    }else {
        $('.show-btn').prop('disabled', true);
        $('.remove-btn').prop('disabled', true);
    }
}

function setDataTableButtonStyles(){
    $('.edit-btn').removeClass("btn-secondary");
    $('.edit-btn').addClass("btn-primary");
    $('.show-btn').removeClass("btn-secondary");
    $('.show-btn').addClass("btn-info");
    $('.remove-btn').removeClass("btn-secondary");
    $('.remove-btn').addClass("btn-danger");
}
function setFormUpdateListener(){
    $("[name='enabled']").bootstrapSwitch();
}
function setAddForm(){
    var defaultNullOptionsData =  {
        "answer_id": 0,
        "answer": "",
        "correct": false
    };
    var defaultOptionsConfig = {
        className: '.add-options',
        formName: 'options[]',
        checkboxName: 'optionsCorrect[]',
        enabledAction: false,
        enabledDeleteOptions: true
    };
    appendOptions(defaultNullOptionsData, defaultOptionsConfig);

    $('#add-options-button').click(function(){
        appendOptions(defaultNullOptionsData, defaultOptionsConfig);
    });
    $('#clear-options-button').click(function(){
        $('.add-options').empty();
        appendOptions(defaultNullOptionsData, defaultOptionsConfig);
    });
    $('#add-question').on('submit', function(e){
        e.preventDefault();
        var question = $('#add-question-text').val();
        var optionsData = $("input[name='options[]']").map(function(){return $(this).val();}).get();
        var correctData = $("input[name='optionsCorrect[]']").map(function(){return $(this).is(":checked")}).get();

        var answers = [];
        for(let i=0; i<optionsData.length; i++){
            answers.push({
                answer: optionsData[i],
                correct: correctData[i]
            });
        }
        ajaxInsertQuestion(question, answers);
    });
}
function setConfirmForm(){
    $('#question-destroyer').click(function(){
        var questionId = $('input[name=confirm-id]').val();
        if(questionId !== ""){
            ajaxDeleteQuestion(questionId);
        }else{
            $('#modal-delete').modal('hide');
        }
    });
}

/* Ajax Section */
function ajaxInsertQuestion(question, answers) {
    var request = {
        question: question,
        answers: answers
    };
    var url = "/api/admin/quiz";
    $.ajax(url, {
        type: 'POST',
        dataType: 'json',
        contentType: 'application/json',
        data: JSON.stringify(request),
        success: function(response, textStatus, jqXHR) {
            Toast.fire({
                type: 'success',
                title: response.message
            });

            Table.ajax.reload();
        }
    });
}
function ajaxDetailData(questionId){
    $('input[name="enabled"]').unbind('switchChange.bootstrapSwitch');
    $('#update-question').unbind('submit');
    resetOptions();

    var url = "/api/admin/quiz/detail?question_id=" + questionId;
    $.ajax(url, {
        dataType: 'json',
        type: 'GET',
        success: function(response, textStatus, jqXHR){
            var result = response.data;

            // populate data
            $("#detail-status").prop('checked', result.active).change();
            $("textarea#detail-question").val(result.question);
            result.answers.forEach(function(opt){
                appendOptions(opt, {
                    className: '.detail-options',
                    enabledAction: true,
                    actionCallback: function(editMode, value, checked){
                        if(!editMode){
                            ajaxUpdateAnswer(opt.answer_id, value, checked);
                        }
                    },
                    defaultFormEnabled: false,
                    enabledDeleteOptions:true,
                    removeCallback: function(value){
                        ajaxDeleteAnswer(opt.answer_id);
                    }
                });
            });
            addSwitchEnabledListener(result.id);

        }
    });
}
function ajaxUpdateQuestion(questionId, question) {
    var request = {
        question: question
    };
    var url = "/api/admin/quiz/update?question_id=" + questionId;
    $.ajax(url, {
        type: 'PUT',
        dataType: 'json',
        contentType: 'application/json',
        data: JSON.stringify(request),
        success: function(response, textStatus, jqXHR) {
            Toast.fire({
                type: 'success',
                title: response.message
            });
        }
    });
}
function ajaxDeleteQuestion(questionId){
    var url = "/api/admin/quiz/delete?question_id=" + questionId;
    $.ajax(url, {
        type: 'DELETE',
        dataType: 'json',
        contentType: 'application/json',
        success: function(response, textStatus, jqXHR) {
            Toast.fire({
                type: 'success',
                title: response.message
            });
            $('#modal-delete').modal('hide');
            Table.ajax.reload();
        }
    });
}
function ajaxUpdateAnswer(answerId, answer, correct) {
    var request = {
        answer: answer,
        correct: correct
    };
    var url = "/api/admin/answer/update?answer_id=" + answerId;
    $.ajax(url, {
        type: 'PUT',
        dataType: 'json',
        contentType: 'application/json',
        data: JSON.stringify(request),
        success: function(response, textStatus, jqXHR) {

            Toast.fire({
                type: 'success',
                title: response.message
            });
        }
    });
}
function ajaxDeleteAnswer(answerId){
    var url = "/api/admin/answer/delete?answer_id=" + answerId;
    $.ajax(url, {
        type: 'DELETE',
        dataType: 'json',
        contentType: 'application/json',
        success: function(response, textStatus, jqXHR) {

            Toast.fire({
                type: 'success',
                title: response.message
            });
        }
    });
}



function resetOptions(){
    $('.detail-options').empty();
}

/*
    options: {
        className,
        checkboxName,
        formName,
        defaultFormEnabled,
        actionCallback,
        enabledAction,
        enabledDeleteOptions,
        removeCallback,
    }
 */
function appendOptions(optionData, options){
    var elementRoot = document.createElement("div");
    elementRoot.className = "form-group";

    var elementOpt = document.createElement("div");
    elementOpt.className = 'input-group input-group-sm';

    // add on icon
    var elementStatus = document.createElement("div");
    elementStatus.className = 'input-group-prepend';
    var elementStatusSpan = document.createElement("span");
    elementStatusSpan.className = 'input-group-text';
    var elementStatusIcon = document.createElement("input");
    elementStatusIcon.type = "checkbox";
    if(typeof options.checkboxName !== 'undefined'){
        elementStatusIcon.name = options.checkboxName;
    }
    elementStatusIcon.checked = optionData.correct;
    if(typeof options.defaultFormEnabled !== 'undefined' && !options.defaultFormEnabled){
        elementStatusIcon.disabled = true;
    }
    elementStatusSpan.append(elementStatusIcon);
    elementStatus.append(elementStatusSpan);
    elementOpt.append(elementStatus);

    // add form
    var formOption = document.createElement("input");
    formOption.type = "text";
    formOption.className = "form-control";
    if(typeof options.formName !== 'undefined'){
        formOption.name = options.formName;
    }
    formOption.value = optionData.answer;
    if(typeof options.defaultFormEnabled !== 'undefined' && !options.defaultFormEnabled){
        formOption.disabled = true;
    }
    elementOpt.append(formOption);

    var elementAction = document.createElement("span");
    elementAction.className = "input-group-append";

    // add button edit
    if(options.enabledAction){
        var elementButton = document.createElement("button");
        elementButton.type = "button";
        elementButton.className = "btn btn-info btn-flat";
        elementButton.innerHTML = "<i class='fas fa-edit'></i>";
        elementButton.onclick = function(){
            formOption.disabled = !formOption.disabled;
            elementStatusIcon.disabled = !elementStatusIcon.disabled;
            editMode = formOption.disabled === false && elementStatusIcon.disabled === false;
            options.actionCallback(editMode, formOption.value, elementStatusIcon.checked);
        };
        elementAction.append(elementButton);
    }

    if(typeof options.enabledDeleteOptions !== 'undefined' && options.enabledDeleteOptions){
        var deleteButton = document.createElement("button");
        deleteButton.type = "button";
        deleteButton.className = "btn btn-danger btn-flat";
        deleteButton.innerHTML = "<i class='fas fa-times'></i>";
        deleteButton.onclick = function(){
            elementRoot.remove();
            options.removeCallback(formOption.value);
        };
        elementAction.append(deleteButton);
    }

    elementOpt.append(elementAction);
    elementRoot.append(elementOpt);

    $(options.className).append(elementRoot);
}


function addSwitchEnabledListener(questionId){
    $('input[name="enabled"]').on('switchChange.bootstrapSwitch', function(event, state) {
        var request = {
            enabled: state
        };
        var url = "/api/admin/quiz/status?question_id=" + questionId;
        $.ajax(url, {
            type:'PUT',
            dataType: 'JSON',
            contentType: 'application/json',
            data: JSON.stringify(request),
            success: function(response, textStatus, jqXHR){

                Toast.fire({
                    type: 'success',
                    title: response.message
                });

            }
        })
    });

    $('#update-question').on('submit', function(e){
       e.preventDefault();

       var question = document.getElementById('detail-question').value;
       ajaxUpdateQuestion(questionId, question);
    });
}