
var gLanguages = [
	"en", ["English", "English"],
	"de", ["Deutsch", "German"],
	"es", ["Español", "Spanish"],
	"fr", ["Français", "French"],
	"pt", ["Português", "Portuguese"],
	"ru", ["Pусский", "Russian"],
	"ro", ["Română", "Romanian"],
	"ja", ["日本語", "Japanese"],
	"cn", ["简体中文", "Chinese"]
];

var gTransalatedPages = [
	"download-free-pdf-viewer", ["cn", "de", "es", "fr", "ja", "pt", "ro"],
	"download-prev", ["de", "es", "ja", "pt", "ro"],
	"downloadafter", ["de", "es", "ja", "pt", "ro"],
	"free-pdf-reader", ["cn", "de", "es", "ja", "pt", "ro"],
	"manual", ["cn", "de", "es", "ja", "pt", "ro", "ru"]
];

// A heuristic used to detect preffered language of the user
// based on settings in navigator object.
// TODO: we should also look at Accept-Language header, which might look like:
// Accept-Language:ru,en-US;q=0.8,en;q=0.6
// but headers are not available from JavaScript so I would have to
// pass this data somehow from the server to html or use additional
// request from javascript as described at http://stackoverflow.com/questions/1043339/javascript-for-detecting-browser-language-preference
function detectBrowserLang() {
	var n = window.navigator;
	// a heuristic: userLanguage and browserLanguage are for ie
	// language is for FireFox and Chrome
	var lang1 = n.userLanguage || n.browserLanguage || n.language || "en";
	// we only care about "en" part of languages like "en-US"
	return lang1.substring(0,2);
}

function installerHref(ver) {
	return '<a href="http://kjkpub.s3.amazonaws.com/sumatrapdf/rel/SumatraPDF-' + ver + '-install.exe">SumatraPDF-' + ver + '-install.exe</a>';
}
function zipHref(ver) {
	return '<a href="http://kjkpub.s3.amazonaws.com/sumatrapdf/rel/SumatraPDF-' + ver + '.zip">SumatraPDF-' + ver + '.zip</a>';
}

// used by download-prev* pages
// Update after releasing a new version
var gPrevSumatraVersion = [
	"1.6", 
	"1.5.1", "1.5", "1.4", "1.3", "1.2", "1.1", "1.0.1",
	"1.0", "0.9.4", "0.9.3", "0.9.1", "0.9", "0.8.1", 
	"0.8", "0.7", "0.6", "0.5", "0.4", "0.3", "0.2"
];

// used by download-prev* pages
function prevLanguagesList(installerStr, zipFileStr) {
	var s, i;
	var s = "";
	for (i=0; i < gPrevSumatraVersion.length; i++)
	{
		var ver = gPrevSumatraVersion[i];
		s += '<p>' + installerStr + ': ' + installerHref(ver) + '<br>\n';
		s += zipFileStr + ': ' + zipHref(ver) + '</p>\n';
	}
	return s;        
}
