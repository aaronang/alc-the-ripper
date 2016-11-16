#!/usr/bin/env ruby

require 'json'
require 'csv'
require 'optparse'
require 'pp'

options = {input: nil, output: nil}
OptionParser.new do |opts|
  opts.banner = "Usage: example.rb [options]"

  opts.on("-i", "--input input", "Input file") do |input|
    options[:input] = input
    options[:output] = input.sub(".json", "_scaling.csv")
  end

  opts.on("-o", "--output output", "Output file") do |output|
    options[:output] = output
  end
end.parse!

file = File.read(options[:input])
data = JSON.parse(file)

CSV.open(options[:output], "w") do |csv|
  csv << ["Time", "Required", "Available", "Actual required"]
  time = 0
  data.each do |o|
    actual_required = if o["jobs"].nil?
      0
    else
      o["jobs"].inject(0) { |sum, j| sum += j["tasks"].size }
    end
    csv << [time, o["requiredSlots"], o["availableSlots"], actual_required]
    time += 10
  end
end
