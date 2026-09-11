Feature: The generated application works through a browser

  These scenarios are part of the generated project. They run unchanged against
  its development server and its standalone production binary.

  Scenario: Go renders a useful initial document
    Given I open the generated application
    Then the heading is "gimble" and the initial Go status is visible

  Scenario: A greeting refreshes the visible status without a reload
    Given I open the generated application
    When I greet "release-qualified"
    Then exactly one greeting from "release-qualified" is visible without a document reload

  Scenario: Kit navigation and a direct deep link both reach About
    Given I open the generated application
    When I follow the About link
    Then About is visible without a document reload
    When I load the About route directly
    Then About is visible in a new document
