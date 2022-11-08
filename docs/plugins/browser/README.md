# browser
Browser plugin is used to interact with a browser
## Available actions
### wait
Waits for a duration
#### Parameters
- duration: Duration to wait
### navigateTo
Navigates to a URL
#### Parameters
- url: URL to navigate to
### urlShouldBe
Checks if the current URL is the expected one
#### Parameters
- url: Expected URL
### textShouldBe
Checks if the text of an element is the expected one
#### Parameters
- selector: Selector of the element
-  (optional) selectorBy: Selector type
- text: Expected text
### sendKeys
Sends keys to an element
#### Parameters
- selector: Selector of the element
-  (optional) selectorBy: Selector type
- keys: Keys to send
### waitVisible
Waits for an element to be visible
#### Parameters
- selector: Selector of the element
-  (optional) selectorBy: Selector type
### click
Clicks on an element
#### Parameters
- selector: Selector of the element
-  (optional) selectorBy: Selector type
### setViewPort
Sets the viewport size
#### Parameters
- width: Width of the viewport
- height: Height of the viewport
### onClose
Close the connection
#### Parameters
