import { Injectable } from '@angular/core';
import 'rxjs/add/operator/toPromise';
import { Router, CanActivate } from '@angular/router';
import {reject, resolve} from 'q';
import { APIService } from "../api.service";

@Injectable()
export class AuthService implements CanActivate {

  constructor(
    private apiService: APIService,
    private router: Router
  ) {}

  public canActivate(): boolean {
    if (this.apiService.isAuthenticated()) {
      return true;
    }
    this.router.navigate(['/']);
    return false;
  }

  public authenticate(user): Promise<any> {
    return this.apiService.post(
        '/v1/auth',
        user
      ).then(
        authResponse => {
          this.apiService.setAuthorization(authResponse.json().authToken);
          this.router.navigate(['']);
        }
      );
  }
}

class AuthRes {
    username: string;
    authToken: string;
    exp: string;
}
