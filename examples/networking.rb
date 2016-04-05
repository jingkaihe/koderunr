require "net/http"
require "uri"

uri = URI.parse("http://google.com/")

# Shortcut
response = Net::HTTP.get_response(uri)

puts response.body
