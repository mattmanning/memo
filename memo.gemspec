# -*- encoding: utf-8 -*-
$:.push File.expand_path("../lib", __FILE__)
require "memo/version"

Gem::Specification.new do |s|
  s.name        = "memo"
  s.version     = Memo::VERSION
  s.authors     = ["Matt Manning"]
  s.email       = ["matt.manning@gmail.com"]
  s.homepage    = "https://github.com/mattmanning/memo"
  s.summary     = %q{A memo pad to store your ideas}
  s.description = %q{A memo pad to store your ideas}

  s.rubyforge_project = "memo"

  s.files         = `git ls-files`.split("\n")
  s.test_files    = `git ls-files -- {test,spec,features}/*`.split("\n")
  s.executables   = `git ls-files -- bin/*`.split("\n").map{ |f| File.basename(f) }
  s.require_paths = ["lib"]

  # specify any dependencies here; for example:
  # s.add_development_dependency "rspec"
  # s.add_runtime_dependency "rest-client"
end
