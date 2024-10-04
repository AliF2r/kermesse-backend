CREATE TYPE users_role_enum AS ENUM ('PARENT', 'STUDENT', 'ORGANIZER', 'STAND_HOLDER');
CREATE TYPE stand_participation_category_enum AS ENUM ('FOOD', 'GAME');
CREATE TYPE started_finished_status_enum AS ENUM ('STARTED', 'FINISHED');

CREATE TABLE "users" (
                         "id" SERIAL PRIMARY KEY,
                         "parent_id" INTEGER REFERENCES "users"("id") DEFAULT NULL,
                         "name" VARCHAR(255) NOT NULL,
                         "email" VARCHAR(255) UNIQUE NOT NULL,
                         "balance" INTEGER NOT NULL DEFAULT 0,
                         "password" VARCHAR(255) NOT NULL,
                         "role" users_role_enum NOT NULL
);

CREATE TABLE "stands" (
                          "id" SERIAL PRIMARY KEY,
                          "user_id" INTEGER NOT NULL UNIQUE REFERENCES "users"("id"),
                          "name" VARCHAR(255) NOT NULL,
                          "category" stand_participation_category_enum NOT NULL,
                          "stock" INTEGER NOT NULL DEFAULT 0,
                          "price" INTEGER NOT NULL DEFAULT 0,
                          "description" TEXT DEFAULT ''
);

CREATE TABLE "kermesses" (
                             "id" SERIAL PRIMARY KEY,
                             "user_id" INTEGER NOT NULL REFERENCES "users"("id"),
                             "name" VARCHAR(255) NOT NULL,
                             "status" started_finished_status_enum NOT NULL DEFAULT 'STARTED',
                             "description" TEXT DEFAULT ''
);

CREATE TABLE "kermesses_users" (
                                   "id" SERIAL PRIMARY KEY,
                                   "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
                                   "user_id" INTEGER NOT NULL REFERENCES "users"("id"),
                                   UNIQUE ("kermesse_id", "user_id")
);

CREATE TABLE "kermesses_stands" (
                                    "id" SERIAL PRIMARY KEY,
                                    "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
                                    "stand_id" INTEGER NOT NULL REFERENCES "stands"("id"),
                                    UNIQUE ("kermesse_id", "stand_id")
);

CREATE TABLE "participations" (
                                "id" SERIAL PRIMARY KEY,
                                "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
                                "stand_id" INTEGER NOT NULL REFERENCES "stands"("id"),
                                "user_id" INTEGER NOT NULL REFERENCES "users"("id"),
                                "category" stand_participation_category_enum NOT NULL,
                                "balance" INTEGER NOT NULL DEFAULT 0,
                                "point" INTEGER NOT NULL DEFAULT 0,
                                "status" started_finished_status_enum NOT NULL DEFAULT 'STARTED'
);

CREATE TABLE "tombolas" (
                            "id" SERIAL PRIMARY KEY,
                            "kermesse_id" INTEGER NOT NULL REFERENCES "kermesses"("id"),
                            "prize" VARCHAR(255) NOT NULL,
                            "name" VARCHAR(255) NOT NULL,
                            "price" INTEGER NOT NULL DEFAULT 0,
                            "status" started_finished_status_enum NOT NULL DEFAULT 'STARTED'
);

CREATE TABLE "tickets" (
                           "id" SERIAL PRIMARY KEY,
                           "user_id" INTEGER NOT NULL REFERENCES "users"("id"),
                           "tombola_id" INTEGER NOT NULL REFERENCES "tombolas"("id"),
                           "is_winner" BOOLEAN NOT NULL DEFAULT FALSE
);
