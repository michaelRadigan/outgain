import * as $ from 'jquery'

export default class Github {
  token: string;
  base: string;

  constructor(token: string, base?: string) {
    this.token = token
    this.base = base || 'https://api.github.com';
  }

  private request(method: string, path: string) {
    return $.ajax({
      method: method,
      url: this.base + path,
      headers: {
        Accept: "application/vnd.github.v3+json",
        Authorization: "token " + this.token,
      },
      dataType: 'json',
    })
  }

  public getGists() {
    return this.request("GET", "/gists");
  }
}
