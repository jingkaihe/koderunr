puts RUBY_VERSION

3.times do |i|
  puts "hello world - #{i}"
  STDOUT.flush
  sleep 1
end
