#!/usr/bin/env ruby

require 'json'
require 'csv'
require 'optparse'

options = {input: nil, output: "data.csv"}
OptionParser.new do |opts|
  opts.banner = "Usage: example.rb [options]"

  opts.on("-i", "--input input", "Input file") do |input|
    options[:input] = input
  end

  opts.on("-o", "--output output", "Output file") do |output|
    options[:output] = output
  end
end.parse!

file = File.read(options[:input])
data = JSON.parse(file)

CSV.open(options[:output], "w") do |csv|
  csv << %w(time required_slots available_slots)
  time = 0
  data.each do |o|
    csv << [time, o["requiredSlots"], o["availableSlots"]]
    time += 10
  end
end
