#!/usr/bin/env ruby

require 'json'
require 'csv'
require 'optparse'

options = {input: nil, output: nil}
OptionParser.new do |opts|
  opts.banner = "Usage: example.rb [options]"

  opts.on("-i", "--input input", "Input file") do |input|
    options[:input] = input
    options[:output] = input.sub(".json", "_balancing.csv")
  end

  opts.on("-o", "--output output", "Output file") do |output|
    options[:output] = output
  end
end.parse!

file = File.read(options[:input])
data = JSON.parse(file)

CSV.open(options[:output], "w") do |csv|
  csv << ["Average tasks running"]
  data.reject{ |o| o["slaves"].nil? }.flat_map do |o|
    o["slaves"].map do |s|
      num_of_tasks = s["tasks"].nil? ? 0 : s["tasks"].size
      { s["ip"] => { tasks: num_of_tasks, samples: 1 } }
    end
  end.inject do |data, sample|
    data.merge(sample) do |k, v1, v2|
      tasks = v1[:tasks] + v2[:tasks]
      samples = v1[:samples] + v2[:samples]
      { tasks: tasks, samples: samples }
    end
  end.map do |_, v|
    average = v[:tasks] / v[:samples].to_f
  end.sort.reverse.each { |avg| csv << [avg] }
end
