import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import {Observable} from 'rxjs/Observable';


export class Application{
  id : number;
  domain: string;
  name: string;
  type: string;
  language: string;
  repositoryurl: string;
  avatarurl: string;
  description: string;
}

class ApplicationListResponse{
  _links : {}
  start: number;
  size: number;
  limit: number;
  results: Application[];
}

@Injectable()
export class ApplicationsService {
  constructor(private http: HttpClient) { }
  listApplications(): Observable<ApplicationListResponse[]> {
    return this.http.get<ApplicationListResponse[]>('api/v0/applications');
  }
  getApplication(app_id): Observable<Application> {
    return this.http.get<Application>('api/v0/applications/' + app_id);
  }
}