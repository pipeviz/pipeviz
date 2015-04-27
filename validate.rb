require 'json-schema'
[1, 2, 3, 4, 5, 6, 7, 8].each do |num|
  res = JSON::Validator.fully_validate('schema.json', "fixtures/ein/#{num}.json")
  unless res.empty?
    puts "fails in #{num}.json\n"
    puts res.join("\n")
  end
end
