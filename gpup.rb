class Gpup < Formula
  desc "Upload files to your Google Photos using the Photos Library API"
  homepage "https://github.com/int128/gpup"
  url "https://github.com/int128/gpup/releases/download/{{ env "VERSION" }}/gpup_darwin_amd64.zip"
  version "{{ env "VERSION" }}"
  sha256 "{{ sha256 .darwin_amd64_archive }}"
  def install
    bin.install "gpup"
  end
  test do
    system "#{bin}/gpup --help"
  end
end