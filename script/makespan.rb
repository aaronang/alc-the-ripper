#!/usr/bin/env ruby

require 'json'
require 'csv'
require 'pp'
require 'time'
require 'optparse'

options = {input: nil, output: nil}
OptionParser.new do |opts|
  opts.banner = "Usage: example.rb [options]"

  opts.on("-i", "--input input", "Input file") do |input|
    options[:input] = input
    options[:output] = input.sub(".json", "_makespan.csv")
  end

  opts.on("-o", "--output output", "Output file") do |output|
    options[:output] = output
  end
end.parse!

file = File.read(options[:input])
data = JSON.parse(file)

CSV.open(options[:output], "w") do |csv|
  csv << ["Makespan"]
  makespans = data.last["completedJobs"].map do |o|
    start = Time.parse(o["startTime"])
    finish = Time.parse(o["finishTime"])
    makespan = finish - start
    makespan / 60.0 - 2.5
  end.sort.reverse
  makespans.each { |m| csv << [m] }
end
