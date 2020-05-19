/* globals JSONPath */
/* eslint-disable import/unambiguous */

// Todo: Extract testing example paths/contents and use for a
//         pulldown that can populate examples

// Todo: Make configurable with other JSONPath options

// Todo: Allow source to be treated as an (evaled) JSON object

// Todo: Could add JSON/JS syntax highlighting in sample and result,
//   ideally with a jsonpath-plus parser highlighter as well

const $ = (s) => document.querySelector(s);

const updateResults = () => {
    const jsonSample = $('#jsonSample');
    const reportValidity = () => {
        // Doesn't work without a timeout
        setTimeout(() => {
            jsonSample.reportValidity();
        });
    };
    let json;
    try {
        json = JSON.parse(jsonSample.value);
        jsonSample.setCustomValidity('');
        reportValidity();
    } catch (err) {
        jsonSample.setCustomValidity('Error parsing JSON: ' + err.toString());
        reportValidity();
        return;
    }
    const result = JSONPath.JSONPath({
        path: $('#jsonpath').value,
        json
    });

    $('#results').value = JSON.stringify(result, null, 2);
};

$('#jsonpath').addEventListener('change', () => {
    updateResults();
});

$('#jsonSample').addEventListener('change', () => {
    updateResults();
});
