# FastColly

Will be changed on Colly:

* simplify colly (rm something)
* add fasthttp as httpbackend

Some designs I don't need, so I will remove them, I will also list them.

* fasthttp API is incompatible with net/http
* fasthttp doesn't support HTTP/2.0 and WebSockets
* remove Appengine client
* scrape func requestData rename as requestBody type is []byte

# Base in Colly v1.2.0

Lightning Fast and Elegant Scraping Framework for Gophers

Colly provides a clean interface to write any kind of crawler/scraper/spider.

With Colly you can easily extract structured data from websites, which can be used for a wide range of applications, like data mining, data processing or archiving.

**[GOCOLLY](https://github.com/gocolly/colly) <- Thanks** 