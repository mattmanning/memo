Gem::Specification.new do |s|
   s.name = %q{memo}
   s.version = '0.0.2'
   s.date = %q{2011-01-02}
   s.authors = ['Matt Manning']
   s.email = %q{matt.manning@gmail.com}
   s.summary = %q{A memo pad to store your ideas}
   s.homepage = %q{https://github.com/mattmanning/memo}
   s.description = %q{A memo pad to store your ideas}
   s.files = [
     'bin/memo',
     'lib/memo.rb',
     'lib/memo/base.rb',
     'README.markdown'
   ]
   s.rubyforge_project = 'memo'
   s.has_rdoc = false
   # s.test_files = ['test/unit.rb']
   s.executables        = %w(memo)
   s.default_executable = "memo"
end