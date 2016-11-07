========================================================
``shimgo`` -- Improved Python Text Processing... From Go
========================================================

.. image:: https://travis-ci.org/tychoish/shimgo.svg?branch=master
   :target: https://travis-ci.org/tychoish/shimgo

Overview
--------

Shimgo is a Go package that converts reStructuredText to HTML. It
wraps the Python docutils implementation of reStructuredText, and is
designed for processing larger quantities of text.

Shimgo also includes support for converting AsciiDoc to HTML via a
similar method. 

Shimgo is primarily intended to support AsciiDoc and reStructuredText
functionality in `hugo <http://gohugo.io>`_, but may also be useful in
supporing similar use cases in the future.

Dependencies
------------

Shimgo requires `flask <http://flask.pocoo.org/>`_, and optionally
docutils for rst support. AsciiDoc support is embedded/vendored.

Internally shimgo depends has *no* third party go libraries.

Development
-----------

In the future, go should have fully featured native implementations of
these markup languages, which will be more efficient and have fewer
dependencies than this system. 

Therefore, the specific implementation and solution are shims.

Nevertheless, pull requests for minor improvements and any issue
reports are more than welcome.
