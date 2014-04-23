/*

Just a place for us to put all of our templates 
(before we actually introduce proper HTML templating).

Placed in the main package so our main function (and its called functions) can access
this easily.

*/
package main


// Clicking on the button on this page brings the person to Evernote
// after getting temp request credentials
const FRONTPAGE = `
<html>
<head><title>Front page</title></head>
<body>
<a href="%s">Click here to log in to Evernote</a>
</body>
</html>
`


// Page that congratulates the user for successfully connecting
// Evernote and Pocket
const DONEPAGE = `
<html>
<head><title></title></head>
<body>
<p>Congratulations. You have successfully connected.</p>
</body>
</html>
`
