# Test task: Web application for analyzing web pages

## Objective

The objective is to build a web application that does an analysis of a web-page/URL.

The application should show a form with a text field in which users can type in the URL of the web page to be analyzed. Additionally, to the form, it should contain a button to send a request to the server.

After processing the results should be shown to the user.

Results should contain next information:

-   What HTML version has the document?
-   What is the page title?
-   How many headings of what level are in the document?
-   How many internal and external links are in the document? Are there any inaccessible links and how many?
-   Does the page contain a login form?

In case the URL given by the user is not reachable an error message should be presented to a user. The message should contain the HTTP status code and a useful error description.

## Restrictions

1.  The application should be written in Golang
2.  The application must be put under git control
3.  You can use whatever libraries/tools you want.

## Submission

Please provide the result as a git repo bundled with:

-   A short text document that lists the main steps of building/deploying your solution as well as all assumptions/decisions you made in case of unclear requirements or missing information
-   Suggestions on possible improvements of the application