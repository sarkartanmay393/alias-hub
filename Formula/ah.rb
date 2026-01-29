class Ah < Formula
  desc "The Ultimate Shell Alias Manager"
  homepage "https://github.com/sarkartanmay393/alias-hub"
  url "https://github.com/sarkartanmay393/alias-hub/archive/refs/tags/v1.0.0.tar.gz"
  # sha256 "" # Will be filled by releaser or user can compute it
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "main.go"
  end

  test do
    system "#{bin}/ah", "--help"
  end
end
