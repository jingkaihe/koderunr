puts RUBY_VERSION

num = gets.scan(/\d+/).map(&:to_i)[0]; puts num

num.times do |i|
  puts "hello world - #{i}"
  STDOUT.flush
  sleep 1
end
