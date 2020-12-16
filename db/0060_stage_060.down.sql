drop index image_white_source_remote_id_idx;

create unique index image_white_source_remote_id_idx
    on image (source, remote_id);
