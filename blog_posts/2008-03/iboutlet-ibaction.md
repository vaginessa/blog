Id: 1167
Title: IBOutlet, IBAction
Tags: cocoa
Date: 2008-03-26T11:56:49-07:00
Format: Markdown
--------------
`IBOutlet` - special instance variable that references another object. A
message can be sent through an outlet. Interface Builder recoganizes
them.

`IBAction` - a special method triggered by user-interface objects.
Interface Builder recognizes them.

<code>\
@interface Controller\
{\
 IBOutlet id textField; // links to TextField UI object\
}

- (IBAction)doAction:(id)sender; // e.g. called when button pushed\
</code>
